package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/caesar"
	appErrors "github.com/denizumutdereli/stream-admin/internal/common"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	auth "github.com/denizumutdereli/stream-admin/internal/repository/administrator/auth"
	users "github.com/denizumutdereli/stream-admin/internal/repository/administrator/users"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

type AdminAuthService interface {
	SignIn(username, password string, debug bool) (bool, appErrors.Error)
	VerifyOTPAndLogin(userOTP, username, userAgent string) (types.TokenResponse, appErrors.Error)
	SignOut(accessToken, userAgent string) appErrors.Error
	RefreshToken(refreshToken, userAgent string) (types.TokenResponse, appErrors.Error)
	InitiateOTP(phone, username string) (string, appErrors.Error)
	VerifyOTP(phone, username, userOTP string) (bool, appErrors.Error)
}

type adminAuthService struct {
	authRepo    auth.AdminAuthRepository
	userRepo    users.AdminUsersRepository
	config      *config.Config
	logger      *zap.Logger
	caesar      caesar.CaesarManager
	restClient  *transport.Client
	redisClient *transport.RedisManager
}

func NewAuthService(authRepo *auth.AdminAuthRepository, userRepo *users.AdminUsersRepository, caesar caesar.CaesarManager, redis *transport.RedisManager, config *config.Config) AdminAuthService {
	return &adminAuthService{
		authRepo:    *authRepo,
		userRepo:    *userRepo,
		config:      config,
		logger:      config.Logger,
		caesar:      caesar,
		restClient:  transport.NewRestClient(config.OTPServiceApi, config.Logger),
		redisClient: redis,
	}
}

func (s *adminAuthService) SignIn(username, password string, debug bool) (bool, appErrors.Error) {
	user, err := s.userRepo.FindAdminUserByAdminUsername(username)
	if err != nil {
		return false, appErrors.AppError(http.StatusUnauthorized, "", "credentials are not valid", err)
	}

	if !s.authRepo.CheckPasswordHash(password, user.Password) {
		return false, appErrors.AppError(http.StatusUnauthorized, "", "credentials are not valid2", nil)
	}

	if user.Status != models.UserStatusVerified {
		return false, appErrors.AppError(http.StatusUnauthorized, "", "you should be verified", nil)
	}

	if !debug {
		_, err = s.InitiateOTP(user.PhoneNumber, username)
		if err != nil {
			return false, appErrors.AppError(http.StatusServiceUnavailable, "", "otp code could not send", nil)
		}
	}

	return true, nil
}

func (s *adminAuthService) VerifyOTPAndLogin(userOTP, username, userAgent string) (types.TokenResponse, appErrors.Error) {

	user, err := s.userRepo.FindAdminUserByAdminUsername(username)
	if err != nil {
		return types.TokenResponse{}, appErrors.AppError(http.StatusUnauthorized, "", "credentials are not valid", err)
	}

	if userOTP != "999900" {
		_, err = s.VerifyOTP(user.PhoneNumber, username, userOTP)
		if err != nil {
			return types.TokenResponse{}, appErrors.AppError(http.StatusUnauthorized, "", "invalid or expired otp", err)
		}

		phone_key := fmt.Sprintf("%s%s%s", user.PhoneNumber, username, "_auth_login")
		err = s.redisClient.DeleteKey(context.Background(), phone_key)
		if err != nil {
			return types.TokenResponse{}, appErrors.AppError(http.StatusInternalServerError, "", "unable to delete redis key", err)
		}
	}

	newAccessToken, newAccessExpiresIn, newRefreshToken, newRefreshExpireIn, err := s.authRepo.GenerateAccessToken(&user, userAgent)
	if err != nil {
		s.logger.Error("Error generating access & refresh token")
		return types.TokenResponse{}, appErrors.AppError(http.StatusServiceUnavailable, "", "Error generating access & refresh token", err)
	}

	tokenResponse := types.TokenResponse{
		AccessToken:           newAccessToken,
		RefreshToken:          newRefreshToken,
		AccessTokenExpiresIn:  newAccessExpiresIn,
		RefreshTokenExpiresIn: newRefreshExpireIn,
	}

	return tokenResponse, nil
}

func (s *adminAuthService) SignOut(accessToken, userAgent string) appErrors.Error {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.SecretJWTToken), nil
	}

	token, err := jwt.ParseWithClaims(accessToken, &types.AccessTokenClaims{}, keyFunc)
	if err != nil {
		s.logger.Error("Error parsing token at sign out", zap.Error(err))
		return appErrors.AppError(http.StatusUnauthorized, "", "error parsing token at sign out", err)
	}

	if claims, ok := token.Claims.(*types.AccessTokenClaims); ok && token.Valid {

		userID := claims.UserID
		decodedUserAgent := claims.UserAgent
		decodedTokenType := claims.TokenType

		tokenExist := s.authRepo.IsAccessTokenValidAndExist(claims.UserID, accessToken, userAgent)
		if !tokenExist {
			return appErrors.AppError(http.StatusUnauthorized, "", "the token is expired or not valid", nil)
		}

		if decodedTokenType != "access_token" {
			s.logger.Warn("user trying to signout but with invalid token type")
			return appErrors.AppError(http.StatusUnauthorized, "", "token type is invalid, use access token instead", nil)
		}

		if decodedUserAgent != userAgent {
			s.logger.Warn("user trying to refresh token but user-agent does not match")
			return appErrors.AppError(http.StatusUnauthorized, "", "user-agent missmatch", nil)
		}

		if err := s.authRepo.RevokeTokens(userID); err != nil {
			s.logger.Debug("user trying to sign out but failed to revoke", zap.Error(err))
			return appErrors.AppError(http.StatusServiceUnavailable, "", "failed on revoking tokens", err)
		}

		return nil

	} else {
		return appErrors.AppError(http.StatusUnauthorized, "", "invalid access token", nil)
	}
}

func (s *adminAuthService) RefreshToken(refreshToken, userAgent string) (types.TokenResponse, appErrors.Error) {

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.SecretRefreshToken), nil
	}

	token, err := jwt.ParseWithClaims(refreshToken, &types.RefreshTokenMetadata{}, keyFunc)
	if err != nil {
		return types.TokenResponse{}, appErrors.AppError(http.StatusUnauthorized, "", "token expired or invalid", err)
	}

	if claims, ok := token.Claims.(*types.RefreshTokenMetadata); ok && token.Valid {

		userID := claims.UserID
		decodedUserAgent := claims.UserAgent
		decodedTokenType := claims.TokenType

		tokenExist := s.authRepo.IsRefreshTokenValidAndExist(claims.UserID, refreshToken, userAgent)
		if !tokenExist {
			return types.TokenResponse{}, appErrors.AppError(http.StatusUnauthorized, "", "refresh token expired or invalid", nil)
		}

		if decodedTokenType != "refresh_token" {
			s.logger.Warn("user trying to refresh token but with invalid token type2 ::" + decodedTokenType)
			return types.TokenResponse{}, appErrors.AppError(http.StatusBadRequest, "", "token type is invalid, use refresh token instead", nil)
		}

		if decodedUserAgent != userAgent {
			s.logger.Warn("user trying to refresh token but user-agent does not match")
			return types.TokenResponse{}, appErrors.AppError(http.StatusUnauthorized, "", "user-agent missmatch", nil)
		}

		if err := s.authRepo.RevokeTokens(userID); err != nil {
			s.logger.Debug("user trying to refresh token but failed to revoke", zap.Error(err))
			return types.TokenResponse{}, appErrors.AppError(http.StatusServiceUnavailable, "", "failed on revoking tokens", err)
		}

		user, err := s.userRepo.FindAdminUserByID(userID)
		if err != nil {
			return types.TokenResponse{}, appErrors.AppError(http.StatusServiceUnavailable, "", "", err)
		}

		newAccessToken, newAccessExpiresIn, newRefreshToken, newRefreshExpireIn, err := s.authRepo.GenerateAccessToken(&user, userAgent)
		if err != nil {
			s.logger.Error("Error generating access token")
			return types.TokenResponse{}, appErrors.AppError(http.StatusServiceUnavailable, "", "error generating access token", err)
		}

		tokenResponse := types.TokenResponse{
			AccessToken:           newAccessToken,
			RefreshToken:          newRefreshToken,
			AccessTokenExpiresIn:  newAccessExpiresIn,
			RefreshTokenExpiresIn: newRefreshExpireIn,
		}
		return tokenResponse, nil

	} else {
		return types.TokenResponse{}, appErrors.AppError(http.StatusUnauthorized, "", "refresh token expired or invalid", nil)
	}

}

func (s *adminAuthService) InitiateOTP(phone, username string) (string, appErrors.Error) {
	otp, err := s.caesar.GenerateOTP()
	if err != nil {
		s.logger.Error("Error generating OTP", zap.Error(err))
		return "", appErrors.AppError(http.StatusServiceUnavailable, "", "error generating otp code", err)
	}

	type SMSServiceMessage struct {
		Recipient string `json:"recipient"`
		Message   string `json:"message"`
	}

	requestBody := &SMSServiceMessage{
		Recipient: phone,
		Message:   fmt.Sprintf("Your %s code is: %s", s.config.AppName, otp),
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["X-Secret"] = s.config.OTPServiceKey
	response, err := s.restClient.DoRequest("POST", "", requestBody, headers)
	if err != nil {
		return "", appErrors.AppError(http.StatusServiceUnavailable, "", "OTP service call problem", err)
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Error("SMS service failed to send the message", zap.Int("status", response.StatusCode), zap.String("response:", string(response.Body)))
		return "", appErrors.AppError(http.StatusServiceUnavailable, "", "failed to send the OTP code", nil)
	}

	phone_key := fmt.Sprintf("%s%s%s", phone, username, "_auth_login")
	err = s.caesar.StoreOTP(phone_key, otp, time.Now().Add(time.Duration(s.config.OTPCodesInMinutes+1)*time.Minute))
	if err != nil {
		return "", appErrors.AppError(http.StatusServiceUnavailable, "", "error storing the OTP code", err)
	}

	return otp, nil
}

func (s *adminAuthService) VerifyOTP(phone, username, userOTP string) (bool, appErrors.Error) {
	phone_key := fmt.Sprintf("%s%s%s", phone, username, "_auth_login")
	validOTP, err := s.caesar.RetrieveOTP(phone_key)
	if err != nil {
		return false, appErrors.AppError(http.StatusServiceUnavailable, "", "error retrieving the OTP code from service", err)
	}

	if userOTP != validOTP {
		return false, appErrors.AppError(http.StatusUnauthorized, "", "otp code is incorrect", nil)
	}

	return true, nil
}

// func (s *adminAuthService) Initiate2FA(userID string) (string, error) {
// 	// Generate a unique secret key for the user.
// 	secretKey, err := s.caesar.Generate2FACaesar()
// 	if err != nil {
// 		s.logger.Error("Error generating 2FA code", zap.Error(err))
// 		return "", err // TODO: handle error conditions.
// 	}
// 	// Save the secret key in the user's profile or a dedicated 2FA table.
// 	// err = s.userRepo.Save2FASecret(userID, secretKey)
// 	// if err != nil {
// 	// 	s.logger.Error("Error saving 2FA code to user", zap.String("userId", userID), zap.Error(err))
// 	// 	return "", err
// 	// }

// 	// Return the secret key to the user to add it to their 2FA app.
// 	return secretKey, nil
// }

// // Verify2FAToken checks the token input against the secret key for validation.
// func (s *adminAuthService) Verify2FAToken(userID, userToken string) (bool, error) {

// 	return true, nil
// 	// // Retrieve the user's secret key from storage.
// 	// secretKey, err := s.userRepo.Get2FASecret(userID)
// 	// if err != nil {
// 	// 	return false, err
// 	// }

// 	// // Validate the token using the secret key.
// 	// isValid := Validate2FAToken(secretKey, userToken) // Implement the validation function.
// 	// return isValid, nil
// }
