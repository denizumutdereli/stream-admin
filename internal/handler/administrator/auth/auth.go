package auth

import (
	"io"
	"net/http"
	"strings"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	service "github.com/denizumutdereli/stream-admin/internal/service/administrator/auth"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminAuthHandler interface {
	Login(c *gin.Context)
	VerifyOTPAndLogin(c *gin.Context)
	Logout(c *gin.Context)
	RefreshToken(c *gin.Context)
	VerifyAccount(c *gin.Context)
}

type adminAuthHandler struct {
	authService service.AdminAuthService
	config      *config.Config
	builders    builders.BuilderService
	logger      *zap.Logger
}

func NewAdminAuthHandler(authService *service.AdminAuthService, cfg *config.Config, builders builders.BuilderService) AdminAuthHandler {
	return &adminAuthHandler{authService: *authService, config: cfg, builders: builders, logger: cfg.Logger}
}

func (ac *adminAuthHandler) Login(c *gin.Context) {
	debug := false
	var loginInfo struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Debug    string `json:"debug" binding:""`
	}

	if err := c.ShouldBindJSON(&loginInfo); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}

		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if loginInfo.Debug != "" {
		debug = true
	}

	_, err := ac.authService.SignIn(loginInfo.Username, loginInfo.Password, debug)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "your credentials are valid and otp code sent",
	})
}

func (ac *adminAuthHandler) VerifyOTPAndLogin(c *gin.Context) {
	var loginInfo struct {
		Username string `json:"username" binding:"required"`
		OTP      string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginInfo); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}

		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	userAgent := c.Request.Header.Get("User-Agent")

	responseTokens, err := ac.authService.VerifyOTPAndLogin(loginInfo.OTP, loginInfo.Username, userAgent)

	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	response := types.TokenResponse{
		AccessToken:           responseTokens.AccessToken,
		RefreshToken:          responseTokens.RefreshToken,
		AccessTokenExpiresIn:  responseTokens.AccessTokenExpiresIn,
		RefreshTokenExpiresIn: responseTokens.RefreshTokenExpiresIn,
	}
	
	c.JSON(http.StatusOK, response)
}

func (ac *adminAuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	if authHeader == "" {
		utils.IfErrorExistReturnWithErrorExplanation(c, nil, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	userAgent := c.Request.Header.Get("User-Agent")

	err := ac.authService.SignOut(accessToken, userAgent)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (ac *adminAuthHandler) RefreshToken(c *gin.Context) {
	var tokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&tokenRequest); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}

		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	userAgent := c.Request.Header.Get("User-Agent")

	responseTokens, err := ac.authService.RefreshToken(tokenRequest.RefreshToken, userAgent)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":             responseTokens.AccessToken,
		"refresh_token":            responseTokens.RefreshToken,
		"acces_token_expires_in":   responseTokens.AccessTokenExpiresIn,
		"refresh_token_expires_in": responseTokens.RefreshTokenExpiresIn,
	})
}

func (ac *adminAuthHandler) VerifyAccount(c *gin.Context) {
	var verificationRequest struct {
		VerificationCode string `json:"verification_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&verificationRequest); err != nil {
		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid format", http.StatusBadRequest)
		return
	}
	// err := ac.authService.VerifyAccount(verificationRequest.VerificationCode)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account"})
	// 	return
	// }
	c.JSON(http.StatusOK, gin.H{"message": "Account successfully verified"})
}
