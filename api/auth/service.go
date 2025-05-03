package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/api/common/location"
	"sync-backend/api/common/session"
	sessionModels "sync-backend/api/common/session/model"
	"sync-backend/api/common/token"
	"sync-backend/api/user"
	userModels "sync-backend/api/user/model"
	"sync-backend/arch/config"
	"sync-backend/arch/network"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const EMPTY_PASSWORD_HASH = "$2a$10$Cv/Xb2ykZ9FLmWyB6vaPEueAzA51kkU2GDZj8C4hwgAH3gQhwIo.q"

type AuthService interface {
	SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, network.ApiError)
	Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, network.ApiError)
	GoogleLogin(googleLoginRequest *dto.GoogleLoginRequest) (*dto.GoogleLoginResponse, network.ApiError)
	Logout(userId string) network.ApiError
	ForgotPassword(forgotPasswordRequest *dto.ForgotPassRequest) network.ApiError
	RefreshToken(refreshTokenRequest *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, network.ApiError)
}

type authService struct {
	network.BaseService
	logger          utils.AppLogger
	userService     user.UserService
	locationService location.LocationService
	sessionService  session.SessionService
	tokenService    token.TokenService
}

func NewAuthService(
	config *config.Config,
	userService user.UserService,
	sessionService session.SessionService,
	locationService location.LocationService,
	tokenService token.TokenService,
) AuthService {
	return &authService{
		BaseService:     network.NewBaseService(),
		logger:          utils.NewServiceLogger("AuthService"),
		userService:     userService,
		locationService: locationService,
		sessionService:  sessionService,
		tokenService:    tokenService,
	}
}

func (s *authService) SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, network.ApiError) {
	s.logger.Info("Signing up user with email: %s", signUpRequest.Email)

	user, err := s.userService.CreateUser(signUpRequest.UserName, signUpRequest.Email, signUpRequest.Password, signUpRequest.ProfilePicUrl)
	if err != nil {
		return nil, network.NewInternalServerError("Error creating user", ERR_USER, err)
	}

	token, err := s.tokenService.GenerateTokenPair(user.UserId)
	if err != nil {
		return nil, network.NewInternalServerError("Error generating token", ERR_TOKEN, err)
	}

	deviceInfo := sessionModels.NewDeviceInfo(signUpRequest.DeviceId, signUpRequest.DeviceName, signUpRequest.DeviceType, signUpRequest.DeviceType, signUpRequest.DeviceModel, signUpRequest.DeviceVersion)

	_, err = s.sessionService.CreateSession(user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), *deviceInfo, signUpRequest.UserAgent, signUpRequest.IPAddress)
	if err != nil {
		return nil, network.NewInternalServerError("Error creating session", ERR_SESSION, err)
	}

	signUpResponse := dto.NewSignUpResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
	s.logger.Success("User signed up successfully: %s", signUpRequest.Email)
	return signUpResponse, nil
}

func (s *authService) Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, network.ApiError) {
	s.logger.Info("Logging in user with email: %s", loginRequest.Email)
	user, err := s.userService.FindUserByEmail(loginRequest.Email)
	if err != nil {
		return nil, network.NewInternalServerError("Error finding user", ERR_USER, err)
	}
	if user == nil {
		return nil, network.NewNotFoundError("User not found", nil)
	}

	// check if user hasnt set password
	if user.PasswordHash == EMPTY_PASSWORD_HASH {
		return nil, network.NewBadRequestError("User has not set password", nil)
	}

	err = s.userService.ValidateUserPassword(user, loginRequest.Password)
	if err != nil {
		return nil, network.NewUnauthorizedError("Entered Password is incorrect", err)
	}

	session, err := s.sessionService.GetUserActiveSession(user.UserId)
	if err != nil {
		return nil, network.NewInternalServerError("Error getting user session", ERR_SESSION, err)
	}
	loginHistory := userModels.LoginHistory{
		LoginTime: primitive.NewDateTimeFromTime(time.Now()),
		IpAddress: loginRequest.IPAddress,
		UserAgent: loginRequest.UserAgent,
		Device: userModels.UserDeviceInfo{
			Os:    loginRequest.DeviceType,
			Type:  loginRequest.DeviceType,
			Name:  loginRequest.DeviceName,
			Model: loginRequest.DeviceModel,
		},
	}
	if session != nil {
		deviceInfo := sessionModels.NewDeviceInfo(loginRequest.DeviceId, loginRequest.DeviceName, loginRequest.DeviceType, loginRequest.DeviceType, loginRequest.DeviceModel, loginRequest.DeviceVersion)

		s.sessionService.UpdateSessionInfo(session.SessionID, *deviceInfo, loginRequest.UserAgent, loginRequest.IPAddress)
		loginHistory.SessionId = session.SessionID
		s.userService.UpdateLoginHistory(user.UserId, loginHistory)

		loginResponse := dto.NewLoginResponse(*user.GetUserInfo(), session.Token, session.RefreshToken)
		s.logger.Success("User logged in successfully: %s", loginRequest.Email)
		// update the login history
		return loginResponse, nil
	} else {
		// Create a new session
		token, err := s.tokenService.GenerateTokenPair(user.UserId)
		if err != nil {
			return nil, network.NewInternalServerError("Error generating token", ERR_TOKEN, err)
		}

		deviceInfo := sessionModels.NewDeviceInfo(loginRequest.DeviceId, loginRequest.DeviceName, loginRequest.DeviceType, loginRequest.DeviceType, loginRequest.DeviceModel, loginRequest.DeviceVersion)

		session, err = s.sessionService.CreateSession(
			user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), *deviceInfo, loginRequest.UserAgent, loginRequest.IPAddress)
		loginHistory.SessionId = session.SessionID
		s.userService.UpdateLoginHistory(user.UserId, loginHistory)

		if err != nil {
			return nil, network.NewInternalServerError("Error creating session", ERR_SESSION, err)
		}
		loginResponse := dto.NewLoginResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
		s.logger.Success("User logged in successfully: %s", loginRequest.Email)
		return loginResponse, nil
	}
}

func (s *authService) GoogleLogin(googleLoginRequest *dto.GoogleLoginRequest) (*dto.GoogleLoginResponse, network.ApiError) {
	s.logger.Info("Logging in user with Google")
	user, err := s.userService.FindUserAuthProvider(googleLoginRequest.GoogleIdToken, userModels.GoogleProviderName)
	if err != nil {
		return nil, network.NewInternalServerError("Error finding user", ERR_USER, err)
	}

	deviceInfo := sessionModels.NewDeviceInfo(googleLoginRequest.DeviceId, googleLoginRequest.DeviceName, googleLoginRequest.DeviceType, googleLoginRequest.DeviceType, googleLoginRequest.DeviceModel, googleLoginRequest.DeviceVersion)

	if user == nil {
		s.logger.Debug("User not found, creating new user")
		user, err := s.userService.CreateUserWithGoogleId(googleLoginRequest.Username, googleLoginRequest.GoogleIdToken)
		if err != nil {
			return nil, network.NewInternalServerError("Error creating user", ERR_USER, err)
		}
		token, err := s.tokenService.GenerateTokenPair(user.UserId)
		if err != nil {
			return nil, network.NewInternalServerError("Error generating token", ERR_TOKEN, err)
		}
		_, err = s.sessionService.CreateSession(user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), *deviceInfo, googleLoginRequest.UserAgent, googleLoginRequest.IPAddress)
		if err != nil {
			return nil, network.NewInternalServerError("Error creating session", ERR_SESSION, err)
		}
		loginResponse := dto.NewGoogleLoginResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
		s.logger.Success("User logged in with Google successfully: %s", user.Email)
		return loginResponse, nil
	} else {
		s.logger.Debug("User found, updating session")
		session, err := s.sessionService.GetUserActiveSession(user.UserId)
		if err != nil {
			return nil, network.NewInternalServerError("Error getting user session", ERR_USER, err)
		}
		if session != nil {
			s.sessionService.UpdateSessionInfo(session.SessionID, *deviceInfo, googleLoginRequest.UserAgent, googleLoginRequest.IPAddress)
			loginResponse := dto.NewGoogleLoginResponse(*user.GetUserInfo(), session.Token, session.RefreshToken)
			s.logger.Success("User logged in with Google successfully: %s", user.Email)
			return loginResponse, nil
		} else {
			token, err := s.tokenService.GenerateTokenPair(user.UserId)
			if err != nil {
				return nil, network.NewInternalServerError("Error generating token", ERR_TOKEN, err)
			}
			_, err = s.sessionService.CreateSession(user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), *deviceInfo, googleLoginRequest.UserAgent, googleLoginRequest.IPAddress)
			if err != nil {
				return nil, network.NewInternalServerError("Error creating session", ERR_SESSION, err)
			}
			loginResponse := dto.NewGoogleLoginResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
			s.logger.Success("User logged in with Google successfully: %s", user.Email)
			return loginResponse, nil
		}
	}
}

func (s *authService) Logout(userId string) network.ApiError {
	s.logger.Info("Logging out user with ID: %s", userId)
	session, err := s.sessionService.GetUserActiveSession(userId)
	if err != nil {
		return network.NewInternalServerError("Error getting session", ERR_SESSION, err)
	}
	if session == nil {
		return network.NewInternalServerError("Session not found", ERR_SESSION_NOT_FOUND, nil)
	}
	err = s.sessionService.InvalidateSession(session.SessionID)
	if err != nil {
		return network.NewInternalServerError("Error invalidating session", ERR_SESSION_INVALID, err)
	}
	s.logger.Success("User logged out successfully: %s", userId)
	return nil
}

func (s *authService) ForgotPassword(forgotPasswordRequest *dto.ForgotPassRequest) network.ApiError {
	s.logger.Info("Processing forgot password for email: %s", forgotPasswordRequest.Email)
	user, err := s.userService.FindUserByEmail(forgotPasswordRequest.Email)
	if err != nil {
		return network.NewInternalServerError("error finding user", ERR_USER, err)
	}
	if user == nil {
		return network.NewNotFoundError("User not found", nil)
	}
	s.logger.Success("Password reset email sent successfully to: %s", forgotPasswordRequest.Email)
	return nil
}

func (s *authService) RefreshToken(refreshTokenRequest *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, network.ApiError) {
	s.logger.Info("Refreshing token")
	session, err := s.sessionService.GetSessionByRefreshToken(refreshTokenRequest.RefreshToken)
	if err != nil {
		return nil, network.NewInternalServerError("Error getting session", ERR_SESSION, err)
	}
	if session == nil {
		return nil, network.NewNotFoundError("Session not found", nil)
	}
	if session.IsExpired() {
		return nil, network.NewUnauthorizedError("Session expired", nil)
	}
	token, err := s.tokenService.GenerateTokenPair(session.UserID)
	if err != nil {
		return nil, network.NewInternalServerError("Error generating token", ERR_TOKEN, err)
	}
	_, err = s.sessionService.UpdateSession(session.SessionID, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time())
	if err != nil {
		return nil, network.NewInternalServerError("Error updating session", ERR_SESSION_INVALID, err)
	}
	refreshTokenResponse := dto.NewRefreshTokenResponse(token.AccessToken, token.RefreshToken)
	s.logger.Success("Tokens refreshed successfully")
	return refreshTokenResponse, nil
}
