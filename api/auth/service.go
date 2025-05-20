package auth

import (
	"sync-backend/api/auth/dto"
	sessionModels "sync-backend/api/common/session/model"
	userModels "sync-backend/api/user/model"

	"sync-backend/api/common/session"
	"sync-backend/api/common/token"
	"sync-backend/api/user"
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
	logger         utils.AppLogger
	userService    user.UserService
	sessionService session.SessionService
	tokenService   token.TokenService
}

func NewAuthService(
	config *config.Config,
	userService user.UserService,
	sessionService session.SessionService,
	tokenService token.TokenService,
) AuthService {
	return &authService{
		BaseService:    network.NewBaseService(),
		logger:         utils.NewServiceLogger("AuthService"),
		userService:    userService,
		sessionService: sessionService,
		tokenService:   tokenService,
	}
}

func (s *authService) SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, network.ApiError) {
	s.logger.Info("Signing up user with email: %s", signUpRequest.Email)

	user, err := s.userService.CreateUser(signUpRequest.UserName, signUpRequest.Email, signUpRequest.Password, signUpRequest.ProfileFilePath, signUpRequest.BackgroundFilePath, signUpRequest.Locale, signUpRequest.TimeZone, signUpRequest.Country)
	if err != nil {
		return nil, NewUserError("creating user", err.Error())
	}

	token, err := s.tokenService.GenerateTokenPair(user.UserId)
	if err != nil {
		return nil, NewTokenError("generating token", err.Error())
	}

	deviceInfo := sessionModels.DeviceInfo{
		DeviceId:        signUpRequest.DeviceId,
		DeviceName:      signUpRequest.DeviceName,
		DeviceType:      signUpRequest.DeviceType,
		DeviceOS:        signUpRequest.DeviceOS,
		DeviceModel:     signUpRequest.DeviceModel,
		DeviceVersion:   signUpRequest.DeviceVersion,
		DeviceUserAgent: signUpRequest.DeviceUserAgent,
	}

	locationInfo := sessionModels.LocationInfo{
		Country:    signUpRequest.Country,
		City:       signUpRequest.City,
		Latitude:   signUpRequest.Latitude,
		Longitude:  signUpRequest.Longitude,
		LocaleCode: signUpRequest.Locale,
		Timezone:   signUpRequest.TimeZone,
		GmtOffset:  signUpRequest.GMTOffset,
		IpAddress:  signUpRequest.IpAddress,
	}

	_, err = s.sessionService.CreateSession(user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), deviceInfo, locationInfo)
	if err != nil {
		return nil, NewSessionError("creating session", err.Error())
	}

	signUpResponse := dto.NewSignUpResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
	s.logger.Success("User signed up successfully: %s", signUpRequest.Email)
	return signUpResponse, nil
}

func (s *authService) Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, network.ApiError) {
	s.logger.Info("Logging in user with email: %s", loginRequest.Email)
	user, err := s.userService.FindUserByEmail(loginRequest.Email)
	if err != nil {
		return nil, NewUserError("finding user", err.Error())
	}
	if user == nil {
		return nil, NewUserNotFoundError(loginRequest.Email)
	}

	// check if user hasnt set password
	if user.PasswordHash == EMPTY_PASSWORD_HASH {
		return nil, NewUserNoPasswordError(loginRequest.Email)
	}

	if user.Status == userModels.Deleted {
		return nil, NewUserDeletedError(loginRequest.Email)
	} else if user.Status == userModels.Banned {
		return nil, NewUserBannedError(loginRequest.Email)
	}

	err = s.userService.ValidateUserPassword(user, loginRequest.Password)
	if err != nil {
		return nil, NewInvalidPasswordError(loginRequest.Email)
	}

	session, err := s.sessionService.GetUserActiveSession(user.UserId)
	if err != nil {
		return nil, NewSessionError("getting user session", err.Error())
	}
	loginHistory := userModels.LoginHistory{
		LoginTime: primitive.NewDateTimeFromTime(time.Now()),
		IpAddress: loginRequest.IpAddress,
		UserAgent: loginRequest.DeviceUserAgent,
		Device: userModels.UserDeviceInfo{
			Os:    loginRequest.DeviceType,
			Type:  loginRequest.DeviceType,
			Name:  loginRequest.DeviceName,
			Model: loginRequest.DeviceModel,
		},
	}
	deviceInfo := sessionModels.DeviceInfo{
		DeviceId:        loginRequest.DeviceId,
		DeviceName:      loginRequest.DeviceName,
		DeviceType:      loginRequest.DeviceType,
		DeviceOS:        loginRequest.DeviceOS,
		DeviceModel:     loginRequest.DeviceModel,
		DeviceVersion:   loginRequest.DeviceVersion,
		DeviceUserAgent: loginRequest.DeviceUserAgent,
	}
	locationInfo := sessionModels.LocationInfo{
		Country:    loginRequest.Country,
		City:       loginRequest.City,
		Latitude:   loginRequest.Latitude,
		Longitude:  loginRequest.Longitude,
		LocaleCode: loginRequest.Locale,
		Timezone:   loginRequest.TimeZone,
		GmtOffset:  loginRequest.GMTOffset,
		IpAddress:  loginRequest.IpAddress,
	}
	if session != nil {
		s.sessionService.UpdateSessionInfo(session.SessionID, deviceInfo, locationInfo)
		loginHistory.SessionId = session.SessionID
		s.userService.UpdateLoginHistory(user.UserId, loginHistory)
		loginResponse := dto.NewLoginResponse(*user.GetUserInfo(), session.Token, session.RefreshToken)
		s.logger.Success("User logged in successfully: %s", loginRequest.Email)
		return loginResponse, nil
	} else {
		token, err := s.tokenService.GenerateTokenPair(user.UserId)
		if err != nil {
			return nil, NewTokenError("generating token", err.Error())
		}

		session, err = s.sessionService.CreateSession(
			user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), deviceInfo, locationInfo)
		loginHistory.SessionId = session.SessionID
		s.userService.UpdateLoginHistory(user.UserId, loginHistory)
		if err != nil {
			return nil, NewSessionError("creating session", err.Error())
		}

		loginResponse := dto.NewLoginResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
		s.logger.Success("User logged in successfully: %s", loginRequest.Email)
		return loginResponse, nil
	}
}

func (s *authService) GoogleLogin(googleLoginRequest *dto.GoogleLoginRequest) (*dto.GoogleLoginResponse, network.ApiError) {
	s.logger.Info("Logging in user with Google")
	user, err := s.userService.FindUserAuthProvider(googleLoginRequest.GoogleIdToken, googleLoginRequest.Username, userModels.GoogleProviderName)
	if err != nil {
		return nil, NewUserError("finding user", err.Error())
	}
	loginHistory := userModels.LoginHistory{
		LoginTime: primitive.NewDateTimeFromTime(time.Now()),
		IpAddress: googleLoginRequest.IpAddress,
		UserAgent: googleLoginRequest.DeviceUserAgent,
		Device: userModels.UserDeviceInfo{
			Os:    googleLoginRequest.DeviceType,
			Type:  googleLoginRequest.DeviceType,
			Name:  googleLoginRequest.DeviceName,
			Model: googleLoginRequest.DeviceModel,
		},
		Provider: userModels.GoogleProviderName,
	}
	deviceInfo := sessionModels.DeviceInfo{
		DeviceId:        googleLoginRequest.DeviceId,
		DeviceName:      googleLoginRequest.DeviceName,
		DeviceType:      googleLoginRequest.DeviceType,
		DeviceOS:        googleLoginRequest.DeviceOS,
		DeviceModel:     googleLoginRequest.DeviceModel,
		DeviceVersion:   googleLoginRequest.DeviceVersion,
		DeviceUserAgent: googleLoginRequest.DeviceUserAgent,
	}
	locationInfo := sessionModels.LocationInfo{
		Country:    googleLoginRequest.Country,
		City:       googleLoginRequest.City,
		Latitude:   googleLoginRequest.Latitude,
		Longitude:  googleLoginRequest.Longitude,
		LocaleCode: googleLoginRequest.Locale,
		Timezone:   googleLoginRequest.TimeZone,
		GmtOffset:  googleLoginRequest.GMTOffset,
		IpAddress:  googleLoginRequest.IpAddress,
	}
	var loginResponse *dto.GoogleLoginResponse
	var session *sessionModels.Session
	if user == nil {
		s.logger.Debug("User not found, creating new user")
		user, err = s.userService.CreateUserWithGoogleId(googleLoginRequest.Username, googleLoginRequest.GoogleIdToken, googleLoginRequest.Locale, googleLoginRequest.TimeZone, googleLoginRequest.Country)
		if err != nil {
			return nil, NewUserError("creating user with GoogleId", err.Error())
		}
		token, err := s.tokenService.GenerateTokenPair(user.UserId)
		if err != nil {
			return nil, NewTokenError("generating token", err.Error())
		}
		session, err = s.sessionService.CreateSession(user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), deviceInfo, locationInfo)
		if err != nil {
			return nil, NewSessionError("creating session", err.Error())
		}
		loginResponse = dto.NewGoogleLoginResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
		s.logger.Success("User logged in with Google successfully: %s", user.Email)
	} else {
		s.logger.Debug("User found, updating session")
		session, err = s.sessionService.GetUserActiveSession(user.UserId)
		if err != nil {
			return nil, NewSessionError("getting user session", err.Error())
		}

		if user.Status == userModels.Deleted {
			return nil, NewUserDeletedError(user.Email)
		} else if user.Status == userModels.Banned {
			return nil, NewUserBannedError(user.Email)
		}

		if session != nil {
			s.sessionService.UpdateSessionInfo(session.SessionID, deviceInfo, locationInfo)
			loginResponse := dto.NewGoogleLoginResponse(*user.GetUserInfo(), session.Token, session.RefreshToken)
			s.logger.Success("User logged in with Google successfully: %s", user.Email)
			return loginResponse, nil
		} else {
			token, err := s.tokenService.GenerateTokenPair(user.UserId)
			if err != nil {
				return nil, NewTokenError("generating token", err.Error())
			}
			session, err = s.sessionService.CreateSession(user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), deviceInfo, locationInfo)
			if err != nil {
				return nil, NewSessionError("creating session", err.Error())
			}
			loginResponse = dto.NewGoogleLoginResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
			s.logger.Success("User logged in with Google successfully: %s", user.Email)
		}
	}
	loginHistory.SessionId = session.SessionID
	s.userService.UpdateLoginHistory(user.UserId, loginHistory)
	return loginResponse, nil
}

func (s *authService) Logout(userId string) network.ApiError {
	s.logger.Info("Logging out user with ID: %s", userId)
	session, err := s.sessionService.GetUserActiveSession(userId)
	if err != nil {
		return NewSessionError("getting session", err.Error())
	}
	if session == nil {
		return NewSessionNotFoundError(userId)
	}
	err = s.sessionService.InvalidateSession(session.SessionID)
	if err != nil {
		return NewSessionInvalidError(session.SessionID)
	}
	s.logger.Success("User logged out successfully: %s", userId)
	return nil
}

func (s *authService) ForgotPassword(forgotPasswordRequest *dto.ForgotPassRequest) network.ApiError {
	s.logger.Info("Processing forgot password for email: %s", forgotPasswordRequest.Email)
	user, err := s.userService.FindUserByEmail(forgotPasswordRequest.Email)
	if err != nil {
		return NewUserError("finding user", err.Error())
	}
	if user == nil {
		return NewUserNotFoundError(forgotPasswordRequest.Email)
	}
	s.logger.Success("Password reset email sent successfully to: %s", forgotPasswordRequest.Email)
	return nil
}

func (s *authService) RefreshToken(refreshTokenRequest *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, network.ApiError) {
	s.logger.Info("Refreshing token")
	session, err := s.sessionService.GetSessionByRefreshToken(refreshTokenRequest.RefreshToken)
	if err != nil {
		return nil, NewSessionError("getting session by refresh token", err.Error())
	}
	userId, err := s.tokenService.GetUserIdFromToken(refreshTokenRequest.RefreshToken)
	if err != nil {
		return nil, NewTokenError("getting user ID from access token", err.Error())
	}
	var accessToken, refreshToken string
	token, err := s.tokenService.GenerateTokenPair(userId)
	if err != nil {
		return nil, NewTokenError("generating token", err.Error())
	}
	if session != nil {
		_, err = s.sessionService.UpdateSession(session.SessionID, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time())
		if err != nil {
			return nil, NewSessionInvalidError(session.SessionID)
		}
		accessToken = token.AccessToken
		refreshToken = token.RefreshToken
	} else {
		deviceInfo := sessionModels.DeviceInfo{
			DeviceId:        refreshTokenRequest.DeviceId,
			DeviceName:      refreshTokenRequest.DeviceName,
			DeviceType:      refreshTokenRequest.DeviceType,
			DeviceOS:        refreshTokenRequest.DeviceOS,
			DeviceModel:     refreshTokenRequest.DeviceModel,
			DeviceVersion:   refreshTokenRequest.DeviceVersion,
			DeviceUserAgent: refreshTokenRequest.DeviceUserAgent,
		}
		locationInfo := sessionModels.LocationInfo{
			Country:    refreshTokenRequest.Country,
			City:       refreshTokenRequest.City,
			Latitude:   refreshTokenRequest.Latitude,
			Longitude:  refreshTokenRequest.Longitude,
			LocaleCode: refreshTokenRequest.Locale,
			Timezone:   refreshTokenRequest.TimeZone,
			GmtOffset:  refreshTokenRequest.GMTOffset,
			IpAddress:  refreshTokenRequest.IpAddress,
		}
		session, err := s.sessionService.CreateSession(userId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), deviceInfo, locationInfo)
		if err != nil {
			return nil, NewSessionError("creating session", err.Error())
		}
		accessToken = session.Token
		refreshToken = session.RefreshToken
	}
	s.logger.Success("Tokens refreshed successfully")
	return dto.NewRefreshTokenResponse(accessToken, refreshToken), nil
}
