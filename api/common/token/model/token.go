package model

import (
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TokenType defines the type of token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// TokenPair represents both access and refresh tokens
type TokenPair struct {
	AccessToken           string             `json:"access_token"`
	RefreshToken          string             `json:"refresh_token"`
	AccessTokenExpiresIn  primitive.DateTime `json:"access_token_expires_in"`
	RefreshTokenExpiresIn primitive.DateTime `json:"refresh_token_expires_in"`
}

// TokenClaims represents the JWT token claims
type TokenClaims struct {
	UserID string    `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.StandardClaims
}
