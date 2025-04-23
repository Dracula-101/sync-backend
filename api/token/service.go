package token

import (
	"crypto/rsa"
	"fmt"
	"sync-backend/api/token/model"
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
	privateKey         *rsa.PrivateKey
	publicKey          *rsa.PublicKey
	accessTokenExpiry  int64
	refreshTokenExpiry int64
	issuer             string
	audience           string
}

func NewTokenService(config *config.Config) TokenService {
	privateKeyBytes, err := utils.LoadPEMFileInto(config.Auth.JWT.PrivateKeyPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load private key: %v", err))
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		panic(fmt.Sprintf("failed to parse private key: %v", err))
	}

	publicKeyBytes, err := utils.LoadPEMFileInto(config.Auth.JWT.PublicKeyPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load public key: %v", err))
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		panic(fmt.Sprintf("failed to parse public key: %v", err))
	}

	// Parse token expiry durations
	accessTokenExpiry := int64(utils.ParseSafeDuration(config.Auth.JWT.AccessTokenExpiry).Seconds())
	refreshTokenExpiry := int64(utils.ParseSafeDuration(config.Auth.JWT.RefreshTokenExpiry).Seconds())

	return &tokenService{
		privateKey:         privateKey,
		publicKey:          publicKey,
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

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signedToken, err := token.SignedString(s.privateKey)
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
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.publicKey, nil
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("invalid token: %w", err)
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

	// Generate a new token pair
	return s.GenerateTokenPair(claims.UserID)
}
