package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/api/session"
	"sync-backend/api/token"
	"sync-backend/api/user"
	"sync-backend/arch/config"
	"sync-backend/arch/network"
)

type AuthService interface {
	SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, network.ApiError)
	Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, network.ApiError)
}

type authService struct {
	network.BaseService
	userService    user.UserService
	sessionService session.SessionService
	tokenService   token.TokenService
}

func NewAuthService(
	userService user.UserService,
	sessionService session.SessionService,
	tokenService token.TokenService,
	config *config.Config,
) AuthService {
	return &authService{
		BaseService:    network.NewBaseService(),
		userService:    userService,
		sessionService: sessionService,
		tokenService:   tokenService,
	}
}

func (s *authService) SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, network.ApiError) {
	user, err := s.userService.CreateUser(signUpRequest.Email, signUpRequest.Password, signUpRequest.Name, signUpRequest.ProfilePicUrl)
	if err != nil {
		return nil, network.NewInternalServerError("error creating user", err)
	}
	token, err := s.tokenService.GenerateTokenPair(user.ID.Hex())
	if err != nil {
		return nil, network.NewInternalServerError("error generating token", err)
	}

	session, _ := s.sessionService.CreateSession(
		user.ID.Hex(), token.AccessToken, token.RefreshToken, signUpRequest.UserAgent, signUpRequest.IPAddress, token.AccessTokenExpiresIn,
	)
	signUpResponse := dto.NewSignUpResponse(
		user.ID.Hex(), session.ID.Hex(), token.AccessToken, token.RefreshToken,
	)
	return signUpResponse, nil
}

func (s *authService) Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, network.ApiError) {
	user, err := s.userService.FindUserByEmail(loginRequest.Email)
	if err != nil {
		return nil, network.NewInternalServerError("error finding user", err)
	}
	if user == nil {
		return nil, network.NewNotFoundError("user not found", nil)
	}

	err = s.userService.ValidateUserPassword(user, loginRequest.Password)
	if err != nil {
		return nil, network.NewUnauthorizedError("invalid password", err)
	}
	session, err := s.sessionService.GetUserActiveSession(user.ID.Hex())
	if err != nil {
		return nil, network.NewInternalServerError("error getting user session", err)
	}
	if session != nil {
		loginResponse := dto.NewLoginResponse(user.ID.Hex(), session.SessionID, session.Token)
		return loginResponse, nil
	} else {
		// Create a new session
		token, err := s.tokenService.GenerateTokenPair(user.ID.Hex())
		if err != nil {
			return nil, network.NewInternalServerError("error generating token", err)
		}
		session, err := s.sessionService.CreateSession(
			user.ID.Hex(), token.AccessToken, token.RefreshToken, loginRequest.UserAgent, loginRequest.IPAddress, token.AccessTokenExpiresIn,
		)
		if err != nil {
			return nil, network.NewInternalServerError("error creating session", err)
		}
		loginResponse := dto.NewLoginResponse(user.ID.Hex(), session.SessionID, token.AccessToken)
		return loginResponse, nil
	}
}
