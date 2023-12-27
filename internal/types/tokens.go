package types

import (
	"github.com/dgrijalva/jwt-go"
)

type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
}

type AccessTokenClaims struct {
	UserID    string `json:"user_id"`
	UserAgent string `json:"user_agent"`
	UserRole  string `json:"user_role"`
	RoleID    string `json:"role_id"`
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
	jwt.StandardClaims
}

type RefreshTokenMetadata struct {
	UserID    string `json:"user_id"`
	UserAgent string `json:"user_agent"`
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
	jwt.StandardClaims
}

type ContextKey string

const (
	ContextUserIDKey ContextKey = "user_id"
	ContextRoleKey   ContextKey = "user_role"
	ContextUserAgent ContextKey = "user_agent"
)

type TokenMetadata struct {
	Token     string `json:"token"`
	UserAgent string `json:"user_agent"`
	TokenType string `json:"token_type"`
	IssuedAt  int64  `json:"issued_at"`
}
