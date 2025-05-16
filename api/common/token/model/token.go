package model

import (
	"encoding/json"

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

// Valid implements jwt.Claims.
func (c TokenClaims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		return err
	}
	if c.Type != AccessToken && c.Type != RefreshToken {
		return jwt.NewValidationError("invalid token type", jwt.ValidationErrorClaimsInvalid)
	}
	return nil
}

func (c *TokenClaims) GetUserID() string {
	return c.UserID
}
func (c *TokenClaims) UnmarshalJSON(data []byte) error {
	type Alias TokenClaims
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	c.UserID = aux.UserID
	c.Type = aux.Type
	return nil
}
