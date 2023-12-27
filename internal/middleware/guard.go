package middleware

import (
	"net/http"
	"strings"

	contextMessageService "github.com/denizumutdereli/stream-admin/internal/comm/message"

	"github.com/denizumutdereli/stream-admin/internal/config"
	administratorAuthRepo "github.com/denizumutdereli/stream-admin/internal/repository/administrator/auth"
	"github.com/denizumutdereli/stream-admin/internal/types"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GuardMiddleware interface {
	Guard() gin.HandlerFunc
}

type guardMiddleware struct {
	config         *config.Config
	logger         *zap.Logger
	jwtSectret     string
	adminAuthRepo  administratorAuthRepo.AdminAuthRepository
	contextMessage contextMessageService.ContextMessages
}

func NewGuardMiddleware(config *config.Config, authRepo administratorAuthRepo.AdminAuthRepository, contextMessages contextMessageService.ContextMessages) GuardMiddleware {
	return &guardMiddleware{
		config:         config,
		logger:         config.Logger,
		jwtSectret:     config.SecretJWTToken,
		adminAuthRepo:  authRepo,
		contextMessage: contextMessages,
	}
}

func (g *guardMiddleware) Guard() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			g.logger.Debug("No auth token provided")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			return
		}

		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			g.logger.Error("Error parsing user token", zap.String("user_token", authHeader))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		tokenString := splitToken[1]

		token, err := jwt.ParseWithClaims(tokenString, &types.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(g.jwtSectret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(*types.AccessTokenClaims); ok && token.Valid {

			verified := g.adminAuthRepo.IsAccessTokenValidAndExist(claims.UserID, tokenString, claims.UserAgent)

			if !verified {

				// Check if there is a contextual message as a reason
				tokenMessage, err := g.contextMessage.GetContextualMessage(claims.UserID, "token", true)
				if err != nil {
					g.logger.Error("Error getting contextual message for why user is not authenticated", zap.Error(err))
				}

				if tokenMessage.Message != "" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": tokenMessage.Message})
				} else {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				}

				return
			}

			g.adminAuthRepo.HandleUserActivity(claims.UserID)

			c.Set(string(types.ContextUserIDKey), claims.UserID)
			c.Set(string(types.ContextRoleKey), claims.RoleID)
			c.Set(string(types.ContextUserAgent), claims.UserAgent)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Next()
	}
}
