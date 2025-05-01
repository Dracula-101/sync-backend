package token

import (
	"fmt"
	"sync-backend/api/common/token/model"
	"sync-backend/arch/config"
	"sync-backend/utils"
	"time"

	"github.com/golang-jwt/jwt"
)

// TokenService defines the interface for token operations
type TokenService interface {
	GenerateTokenPair(userId string) (*model.TokenPair, error)
	ValidateToken(tokenString string) (*jwt.Token, *model.TokenClaims, error)
	RefreshTokens(refreshToken string) (*model.TokenPair, error)
}

type tokenService struct {
	secretKey          []byte
	accessTokenExpiry  int64
	refreshTokenExpiry int64
	issuer             string
	audience           string
}

func NewTokenService(config *config.Config) TokenService {
	// Load secret key for HMAC-SHA256
	secretKey := []byte(config.Auth.JWT.SecretKey)
	if len(secretKey) == 0 {
		panic("JWT secret key is empty")
	}

	// Parse token expiry durations
	accessTokenExpiry := int64(utils.ParseSafeDuration(config.Auth.JWT.AccessTokenExpiry).Seconds())
	refreshTokenExpiry := int64(utils.ParseSafeDuration(config.Auth.JWT.RefreshTokenExpiry).Seconds())

	return &tokenService{
		secretKey:          secretKey,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		issuer:             config.Auth.JWT.Issuer,
		audience:           config.Auth.JWT.Audience,
	}
}

func (s *tokenService) generateToken(userId string, tokenType model.TokenType, expiresIn int64) (string, error) {
	now := time.Now()
	claims := model.TokenClaims{
		UserID: userId,
		Type:   tokenType,
		StandardClaims: jwt.StandardClaims{
			Subject:   userId,
			Issuer:    s.issuer,
			Audience:  s.audience,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(time.Duration(expiresIn) * time.Second).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (s *tokenService) GenerateTokenPair(userId string) (*model.TokenPair, error) {
	accessToken, err := s.generateToken(userId, model.AccessToken, s.accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(userId, model.RefreshToken, s.refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &model.TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresIn:  time.Now().Add(time.Duration(s.accessTokenExpiry) * time.Second),
		RefreshTokenExpiresIn: time.Now().Add(time.Duration(s.refreshTokenExpiry) * time.Second),
	}, nil
}

func (s *tokenService) ValidateToken(tokenString string) (*jwt.Token, *model.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&model.TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.secretKey, nil
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("invalid token: %s - %w", tokenString, err)
	}

	if !token.Valid {
		return nil, nil, fmt.Errorf("token validation failed")
	}

	claims, ok := token.Claims.(*model.TokenClaims)
	if !ok {
		return nil, nil, fmt.Errorf("invalid token claims format")
	}

	if claims.Issuer != s.issuer {
		return nil, nil, fmt.Errorf("invalid token issuer")
	}

	if claims.Audience != s.audience {
		return nil, nil, fmt.Errorf("invalid token audience")
	}

	return token, claims, nil
}

func (s *tokenService) RefreshTokens(refreshToken string) (*model.TokenPair, error) {
	_, claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Ensure this is actually a refresh token
	if claims.Type != model.RefreshToken {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Generate a new token pair using the same session ID
	return s.GenerateTokenPair(claims.UserID)
}
