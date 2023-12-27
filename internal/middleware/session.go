package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	contextMessageService "github.com/denizumutdereli/stream-admin/internal/comm/message"
	administratorUserService "github.com/denizumutdereli/stream-admin/internal/service/administrator/users"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/go-redis/redis/v8"

	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SessionMiddleware interface {
	RefreshTimeout() gin.HandlerFunc
	CheckUserLock() gin.HandlerFunc
	LimitOTPAttempts() gin.HandlerFunc
	NotDeleteOwnUser() gin.HandlerFunc
}

type sessionMiddleware struct {
	config            *config.Config
	logger            *zap.Logger
	redis             *transport.RedisManager
	adminUsersService administratorUserService.AdminUserService
	contextMessage    contextMessageService.ContextMessages
}

func NewSessionMiddleware(config *config.Config, redis *transport.RedisManager, adminUsersService administratorUserService.AdminUserService, contextMessages contextMessageService.ContextMessages) SessionMiddleware {
	return &sessionMiddleware{
		config:            config,
		logger:            config.Logger,
		redis:             redis,
		adminUsersService: adminUsersService,
		contextMessage:    contextMessages,
	}
}

func (s *sessionMiddleware) RefreshTimeout() gin.HandlerFunc {
	return func(c *gin.Context) {
		s.logger.Debug("session refresh timeout...")
		c.Next()
	}
}

func (s *sessionMiddleware) CheckUserLock() gin.HandlerFunc {
	return func(c *gin.Context) {
		var locked *bool
		ctx := c.Request.Context()

		userID, exists := c.Get(string(types.ContextUserIDKey))
		if !exists {
			s.logger.Error("Username not found in context")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Username required"})
			return
		}

		userIDKey, ok := userID.(string)
		if !ok {
			s.logger.Error("Assertion problem on checking userID on lock")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Assertion problem on checking userID on lock"})
		}

		lockKey := fmt.Sprintf("user-lock:%s", userIDKey)

		err := s.redis.GetKeyValue(ctx, lockKey, &locked)
		if err != nil && err != redis.Nil {
			s.logger.Error("Error checking user lock", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error checking user status"})
			return
		}

		if locked != nil {
			s.logger.Warn("User action locked", zap.String("userID", userIDKey))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User is currently locked out. Try again later."})
			return
		}

		c.Next()
	}
}

func (s *sessionMiddleware) LimitOTPAttempts() gin.HandlerFunc {
	return func(c *gin.Context) {

		type otpRequest struct {
			Username string `json:"username"`
		}

		if c.ContentType() != "application/json" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "'BindWithoutBindingTag' only serves for application/json not for " + c.ContentType()})
			return
		}

		bodyAsByteArray, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error reading request body. ref:middleware"})
			return
		}

		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyAsByteArray))

		var req otpRequest
		err = json.Unmarshal(bodyAsByteArray, &req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		username := req.Username

		if username == "" {
			s.logger.Error("Username not found in context")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Username required in context"})
			return
		}

		otpAttemptsKey := fmt.Sprintf("otp-attempts:%s", username)
		ctx := c.Request.Context()

		var attempts int
		err = s.redis.GetKeyValue(ctx, otpAttemptsKey, &attempts)
		if err == redis.Nil {
			attempts = 0
		} else if err != nil {
			s.logger.Error("Failed to get OTP attempts from Redis", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if attempts >= s.config.OTPCodesMaxTry {
			s.logger.Warn("OTP attempt limit exceeded", zap.String("otpKey", username))
			_, err := s.adminUsersService.SetUserLock(username)
			if err != nil {
				s.logger.Error("Failed to set user lock operation on Redis", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
			s.logger.Debug("User locked", zap.String("user_id", username))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many OTP attempts, please try again later"})
			return
		}

		attempts++
		expiration := time.Minute * time.Duration(s.config.OtpCodesMaxTryInARowInMinutes)
		err = s.redis.SetKeyValue(ctx, otpAttemptsKey, attempts, expiration)
		if err != nil {
			s.logger.Error("Failed to set OTP attempts in Redis", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		s.logger.Debug("There is no problem with OTP attempts", zap.Int("attemtps", attempts))
		c.Next()
	}
}

func (s *sessionMiddleware) NotDeleteOwnUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		s.logger.Debug("checking if user is trying to delete self...")

		sessionUserID, exists := c.Get(string(types.ContextUserIDKey))
		if !exists {
			s.logger.Error("Session user ID not found in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in session"})
			return
		}

		requestUserID := c.Param("user_id")

		if sessionUserID == requestUserID {
			s.logger.Error("User is attempting to delete self")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You cannot delete your own account"})
			return
		}

		c.Next()
	}
}
