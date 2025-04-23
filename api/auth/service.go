package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/api/session"
	"sync-backend/api/token"
	"sync-backend/api/user"
	"sync-backend/arch/config"
	"sync-backend/arch/network"
	"sync-backend/utils"
)

type AuthService interface {
	SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, network.ApiError)
	Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, network.ApiError)
	GoogleLogin(googleLoginRequest *dto.GoogleLoginRequest) (*dto.GoogleLoginResponse, network.ApiError)
	Logout(logoutRequest *dto.LogoutRequest) network.ApiError
}

type authService struct {
	network.BaseService
	logger         utils.AppLogger
	userService    user.UserService
	sessionService session.SessionService
	tokenService   token.TokenService
}

func NewAuthService(
	logger utils.AppLogger,
	userService user.UserService,
	sessionService session.SessionService,
	tokenService token.TokenService,
	config *config.Config,
) AuthService {
	return &authService{
		BaseService:    network.NewBaseService(),
		logger:         logger,
		userService:    userService,
		sessionService: sessionService,
		tokenService:   tokenService,
	}
}

func (s *authService) SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, network.ApiError) {
	s.logger.Info("Signing up user with email: %s", signUpRequest.Email)
	user, err := s.userService.CreateUser(signUpRequest.Email, signUpRequest.Password, signUpRequest.Name, signUpRequest.ProfilePicUrl)
	if err != nil {
		return nil, network.NewInternalServerError("error creating user", err)
	}
	token, err := s.tokenService.GenerateTokenPair(user.ID.Hex())
	if err != nil {
		return nil, network.NewInternalServerError("error generating token", err)
	}

	s.sessionService.CreateSession(
		user.ID.Hex(), token.AccessToken, token.RefreshToken, signUpRequest.UserAgent, signUpRequest.IPAddress, token.AccessTokenExpiresIn,
	)
	signUpResponse := dto.NewSignUpResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
	s.logger.Success("User signed up successfully: %s", signUpRequest.Email)
	return signUpResponse, nil
}

func (s *authService) Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, network.ApiError) {
	s.logger.Info("Logging in user with email: %s", loginRequest.Email)
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
		loginResponse := dto.NewLoginResponse(*user.GetUserInfo(), session.Token)
		s.logger.Success("User logged in successfully: %s", loginRequest.Email)
		return loginResponse, nil
	} else {
		// Create a new session
		token, err := s.tokenService.GenerateTokenPair(user.ID.Hex())
		if err != nil {
			return nil, network.NewInternalServerError("error generating token", err)
		}
		_, err = s.sessionService.CreateSession(
			user.ID.Hex(), token.AccessToken, token.RefreshToken, loginRequest.UserAgent, loginRequest.IPAddress, token.AccessTokenExpiresIn,
		)
		if err != nil {
			return nil, network.NewInternalServerError("error creating session", err)
		}
		loginResponse := dto.NewLoginResponse(*user.GetUserInfo(), token.AccessToken)
		s.logger.Success("User logged in successfully: %s", loginRequest.Email)
		return loginResponse, nil
	}
}

func (s *authService) GoogleLogin(googleLoginRequest *dto.GoogleLoginRequest) (*dto.GoogleLoginResponse, network.ApiError) {
	s.logger.Info("Logging in user with Google")
	user, err := s.userService.GetUserByGoogleId(googleLoginRequest.GoogleIdToken)
	if err != nil {
		return nil, network.NewInternalServerError("error finding user", err)
	}
	if user == nil {
		user, err = s.userService.CreateUserWithGoogleId(googleLoginRequest.GoogleIdToken)
		if err != nil {
			return nil, network.NewInternalServerError("error creating user", err)
		}
	}
	token, err := s.tokenService.GenerateTokenPair(user.ID.Hex())
	if err != nil {
		return nil, network.NewInternalServerError("error generating token", err)
	}
	_, err = s.sessionService.CreateSession(
		user.ID.Hex(), token.AccessToken, token.RefreshToken, googleLoginRequest.UserAgent, googleLoginRequest.IPAddress, token.AccessTokenExpiresIn,
	)
	if err != nil {
		return nil, network.NewInternalServerError("error creating session", err)
	}
	loginResponse := dto.NewGoogleLoginResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
	s.logger.Success("User logged in successfully with Google")
	return loginResponse, nil
}

func (s *authService) Logout(logoutRequest *dto.LogoutRequest) network.ApiError {
	s.logger.Info("Logging out user with ID: %s", logoutRequest.UserId)
	session, err := s.sessionService.GetUserActiveSession(logoutRequest.UserId)
	if err != nil {
		return network.NewInternalServerError("error getting session", err)
	}
	if session == nil {
		return network.NewInternalServerError("session not found", nil)
	}
	err = s.sessionService.InvalidateSession(session.SessionID)
	if err != nil {
		return network.NewInternalServerError("error invalidating session", err)
	}
	s.logger.Success("User logged out successfully: %s", logoutRequest.UserId)
	return nil
}
