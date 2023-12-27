package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	contextMessageService "github.com/denizumutdereli/stream-admin/internal/comm/message"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminAuthRepository interface {
	GenerateAccessToken(user *models.AdministratorUser, userAgent string) (string, int64, string, int64, error)
	GenerateRefreshToken(user *models.AdministratorUser, userAgent string) (string, int64, error)
	RevokeTokens(userID string) error
	RevokeRefreshToken(userID string) error
	GeneratePasswordResetToken(userID string) (string, error)
	SendPasswordResetEmail(email, resetToken string) error // TODO: Rotate in TTL
	IsAccessTokenValidAndExist(userID, accessToken, userAgent string) bool
	IsRefreshTokenValidAndExist(userID, refreshToken, userAgent string) bool
	CheckPasswordHash(password, hash string) bool
	HashPassword(password string) (string, error)
	HandleUserActivity(userID string)
}

type repoConfig struct {
	ServicePrefix                string
	AdministratorAuthConfigTable string
}

type adminAuthRepository struct {
	ctx                context.Context
	cancel             context.CancelFunc
	config             *config.Config
	database           *gorm.DB
	repoConfig         *repoConfig
	mutex              sync.RWMutex
	sessionTimers      map[string]*time.Timer
	sessionTimersMutex sync.Mutex
	logger             *zap.Logger
	builders           builders.BuilderService
	redisClient        *transport.RedisManager
	contextMessage     contextMessageService.ContextMessages
}

func NewAuthRepository(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService, redis *transport.RedisManager, contextMessages contextMessageService.ContextMessages) (AdminAuthRepository, error) {
	database.AutoMigrate(&models.AdministratorAuth{})

	repoConfig := &repoConfig{
		ServicePrefix:                servicePrefix,
		AdministratorAuthConfigTable: servicePrefix + "_auth_config"}

	err := config.PrefixService.RegisterServiceTables(servicePrefix, []string{repoConfig.AdministratorAuthConfigTable})
	if err != nil {
		return nil, err
	}

	repository := &adminAuthRepository{
		database:       database,
		repoConfig:     repoConfig,
		config:         config,
		logger:         config.Logger,
		builders:       builders,
		redisClient:    redis,
		sessionTimers:  make(map[string]*time.Timer),
		contextMessage: contextMessages,
	}

	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (a *adminAuthRepository) GenerateAccessToken(user *models.AdministratorUser, userAgent string) (string, int64, string, int64, error) {
	expirationTime := time.Now().Add(time.Duration(a.config.DefaultPanelAccessTokenTimeOut) * time.Minute)

	roleName := "superAdmin"

	claims := &types.AccessTokenClaims{
		UserID:    user.UserID,
		UserAgent: userAgent,
		UserRole:  roleName,
		RoleID:    user.UserRole,
		TokenType: "access_token",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   user.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.config.SecretJWTToken))
	if err != nil {
		return "", 0, "", 0, err
	}

	redisKey := fmt.Sprintf("access_token:%s", user.UserID)
	tokenData := map[string]interface{}{
		"token":      tokenString,
		"user_agent": userAgent,
		"issued_at":  time.Now().Unix(),
	}

	refreshToken, refreshTokenExpiresIn, err := a.GenerateRefreshToken(user, userAgent)
	if err != nil {
		return "", 0, "", 0, err
	}

	err = a.redisClient.SetKeyValue(a.ctx, redisKey, tokenData,
		time.Duration(a.config.DefaultPanelRefreshTokenTimeOut)*time.Minute)
	if err != nil {
		return "", 0, "", 0, err
	}

	//a.logger.Debug("user timer has started")
	a.resetUserTimer(user.UserID)

	if err != nil {
		a.logger.Error("failed to set context message", zap.Error(err))
	}

	return tokenString, expirationTime.Unix(), refreshToken, refreshTokenExpiresIn, nil
}

func (a *adminAuthRepository) GenerateRefreshToken(user *models.AdministratorUser, userAgent string) (string, int64, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	expirationTime := time.Now().Add(time.Duration(a.config.DefaultPanelRefreshTokenTimeOut) * time.Minute) // TODO: from config

	claims := &types.RefreshTokenMetadata{
		UserID:    user.UserID,
		UserAgent: userAgent,
		TokenType: "refresh_token",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   user.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.config.SecretRefreshToken))
	if err != nil {
		return "", 0, err
	}

	redisKey := fmt.Sprintf("refresh_token:%s", user.UserID)
	tokenData := map[string]interface{}{
		"token":      tokenString,
		"user_agent": userAgent,
		"token_type": "refresh_token",
		"issued_at":  time.Now().Unix(),
	}
	err = a.redisClient.SetKeyValue(a.ctx, redisKey, tokenData,
		time.Duration(a.config.DefaultPanelRefreshTokenTimeOut)*time.Minute)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

func (a *adminAuthRepository) RevokeTokens(userID string) error {
	redisKey := fmt.Sprintf("access_token:%s", userID)

	err := a.redisClient.DeleteKey(a.ctx, redisKey)
	if err != nil {
		return err
	}

	err = a.RevokeRefreshToken(userID)
	if err != nil {
		return err
	}
	return nil
}

func (a *adminAuthRepository) RevokeRefreshToken(userID string) error {
	redisKey := fmt.Sprintf("refresh_token:%s", userID)

	err := a.redisClient.DeleteKey(a.ctx, redisKey)
	return err
}

// Mock
func (a *adminAuthRepository) GeneratePasswordResetToken(userID string) (string, error) {
	resetToken := fmt.Sprintf("reset_%s.%d", userID, time.Now().UnixNano())

	expiration := time.Duration(1) * time.Hour

	err := a.redisClient.SetKeyValue(a.ctx, resetToken, userID,
		expiration) // TODO: config

	if err != nil {
		return "", err
	}

	return resetToken, nil
}

// Mock
func (a *adminAuthRepository) SendPasswordResetEmail(email, resetToken string) error {
	return nil
}

func (a *adminAuthRepository) IsAccessTokenValidAndExist(userID, accessToken, userAgent string) bool {
	var storedAccessToken types.AccessTokenClaims
	redisKey := fmt.Sprintf("access_token:%s", userID)

	err := a.redisClient.GetKeyValue(a.ctx, redisKey, &storedAccessToken)
	if err != nil {
		return false
	}
	isTokenExist := storedAccessToken.Token == accessToken
	isUserAgentMatching := storedAccessToken.UserAgent == userAgent

	return isTokenExist && isUserAgentMatching
}

func (a *adminAuthRepository) IsRefreshTokenValidAndExist(userID, refreshToken, userAgent string) bool {
	var storedRefreshToken types.RefreshTokenMetadata
	redisKey := fmt.Sprintf("refresh_token:%s", userID)

	err := a.redisClient.GetKeyValue(a.ctx, redisKey, &storedRefreshToken)
	if err != nil {
		return false
	}

	isTokenExist := storedRefreshToken.Token == refreshToken
	isUserAgentMatching := storedRefreshToken.UserAgent == userAgent

	return isTokenExist && isUserAgentMatching
}

func (a *adminAuthRepository) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (a *adminAuthRepository) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (a *adminAuthRepository) resetUserTimer(userID string) {
	a.sessionTimersMutex.Lock()
	defer a.sessionTimersMutex.Unlock()

	if timer, exists := a.sessionTimers[userID]; exists {
		timer.Stop()
	}

	//a.logger.Debug("user timer triggered...")

	a.sessionTimers[userID] = time.AfterFunc(time.Duration(a.config.DefaultPanelIdleSessionTimeOut*int(time.Minute)), func() {
		a.logger.Debug("user timout, logging out")
		var redisTimeout int = 1
		err := a.contextMessage.SetContextualMessage(&types.ContextualMessage{
			UserId:                userID,
			MessageType:           "token",
			Message:               "idle session timeout",
			RedisDelivery:         true,
			NatsDelivery:          true,
			RedisTimeoutInMinutes: &redisTimeout,
		})

		if err != nil {
			a.logger.Error("contextual message error", zap.Error(err))
		}

		a.RevokeTokens(userID)
	})
}

func (a *adminAuthRepository) HandleUserActivity(userID string) {
	a.resetUserTimer(userID)
}
