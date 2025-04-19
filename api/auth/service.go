package auth

import (
	"crypto/rsa"
	"fmt"
	"sync-backend/api/auth/dto"
	"sync-backend/api/user"
	"sync-backend/arch/config"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, error)
	Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, error)

	GenerateToken(userId string) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error)
}

type authService struct {
	network.BaseService
	userService user.UserService
	config      *config.Config
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
}

func NewAuthService(
	db mongo.Database,
	userService user.UserService,
	config *config.Config,
) AuthService {
	// Load RSA keys
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

	return &authService{
		BaseService: network.NewBaseService(),
		userService: userService,
		config:      config,
		privateKey:  privateKey,
		publicKey:   publicKey,
	}
}

func (s *authService) SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, error) {
	user, err := s.userService.CreateUser(signUpRequest.Email, signUpRequest.Password, signUpRequest.Name, signUpRequest.ProfilePicUrl)
	if err != nil {
		return nil, err
	}
	token, err := s.GenerateToken(user.ID.Hex())
	if err != nil {
		return nil, err
	}
	signUpResponse := dto.NewSignUpResponse(user.ID.Hex(), token)
	return signUpResponse, nil
}

func (s *authService) Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userService.FindUserByEmail(loginRequest.Email)
	if err != nil {
		return nil, err
	}

	err = s.userService.ValidateUserPassword(user, loginRequest.Password)
	if err != nil {
		return nil, err
	}
	token, err := s.GenerateToken(user.ID.Hex())
	if err != nil {
		return nil, err
	}
	loginResponse := dto.NewLoginResponse(user.ID.Hex(), token)
	return loginResponse, nil
}

func (s *authService) GenerateToken(userId string) (string, error) {
	// Parse the token expiration from config
	accessExpiry, err := utils.ParseDuration(s.config.Auth.JWT.AccessTokenExpiry)
	if err != nil {
		return "", fmt.Errorf("invalid access token expiry configuration: %w", err)
	}

	// Create token claims
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userId,                       // Subject (user ID)
		"iss": s.config.Auth.JWT.Issuer,     // Issuer
		"aud": s.config.Auth.JWT.Audience,   // Audience
		"iat": now.Unix(),                   // Issued at
		"exp": now.Add(accessExpiry).Unix(), // Expiry time
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign the token with our private key
	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, nil, fmt.Errorf("token validation failed")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("invalid token claims")
	}

	// Verify issuer and audience
	if claims["iss"] != s.config.Auth.JWT.Issuer {
		return nil, nil, fmt.Errorf("invalid token issuer")
	}

	if claims["aud"] != s.config.Auth.JWT.Audience {
		return nil, nil, fmt.Errorf("invalid token audience")
	}

	return token, claims, nil
}
