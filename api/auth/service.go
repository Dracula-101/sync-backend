package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync-backend/api/auth/dto"
	"sync-backend/api/common/email"
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
	VerifyEmail(token string) (*userModels.User, network.ApiError)
	ResetPassword(token string, newPassword string) network.ApiError
}

type authService struct {
	network.BaseService
	logger         utils.AppLogger
	config         *config.Config
	env            *config.Env
	userService    user.UserService
	sessionService session.SessionService
	tokenService   token.TokenService
	emailService   email.EmailService
}

func NewAuthService(
	config *config.Config,
	env *config.Env,
	userService user.UserService,
	sessionService session.SessionService,
	tokenService token.TokenService,
	emailService email.EmailService,
) AuthService {
	return &authService{
		BaseService:    network.NewBaseService(),
		logger:         utils.NewServiceLogger("AuthService"),
		config:         config,
		env:            env,
		userService:    userService,
		sessionService: sessionService,
		tokenService:   tokenService,
		emailService:   emailService,
	}
}

func (s *authService) SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, network.ApiError) {
	s.logger.Info("Signing up user with email: %s", signUpRequest.Email)

	user, err := s.userService.CreateUser(signUpRequest.UserName, signUpRequest.Email, signUpRequest.Password, signUpRequest.ProfileFilePath, signUpRequest.BackgroundFilePath, signUpRequest.Locale, signUpRequest.TimeZone, signUpRequest.Country)
	if err != nil {
		return nil, err
	}

	token, tokenErr := s.tokenService.GenerateTokenPair(user.UserId)
	if tokenErr != nil {
		return nil, NewTokenError("generating token", tokenErr.Error())
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

	_, sessionErr := s.sessionService.CreateSession(user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), deviceInfo, locationInfo)
	if sessionErr != nil {
		return nil, NewSessionError("creating session", sessionErr.Error())
	}

	signUpResponse := dto.NewSignUpResponse(*user.GetUserInfo(), token.AccessToken, token.RefreshToken)
	s.logger.Success("User signed up successfully: %s", signUpRequest.Email)
	return signUpResponse, nil
}

func (s *authService) Login(loginRequest *dto.LoginRequest) (*dto.LoginResponse, network.ApiError) {
	s.logger.Info("Logging in user with email: %s", loginRequest.Email)
	user, err := s.userService.FindUserByEmail(loginRequest.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, NewUserNotFoundError(loginRequest.Email)
	}

	// check if user hasnt set password
	if user.PasswordHash == EMPTY_PASSWORD_HASH {
		return nil, NewUserNoPasswordError(loginRequest.Email)
	}

	switch user.Status {
	case userModels.Deleted:
		return nil, NewUserDeletedError(loginRequest.Email)
	case userModels.Banned:
		return nil, NewUserBannedError(loginRequest.Email, "Banned due to violation of terms of service")
	}

	err = s.userService.ValidateUserPassword(user, loginRequest.Password)
	if err != nil {
		return nil, err
	}

	session, sessionErr := s.sessionService.GetUserActiveSession(user.UserId)
	if sessionErr != nil {
		return nil, NewSessionError("getting user session", sessionErr.Error())
	}
	loginHistory := userModels.LoginHistory{
		LoginTime: primitive.NewDateTimeFromTime(time.Now()),
		IpAddress: loginRequest.IpAddress,
		UserAgent: loginRequest.DeviceUserAgent,
		Device: userModels.UserDeviceInfo{
			Os:    loginRequest.DeviceOS,
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
		return nil, err
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
		session, sessionErr := s.sessionService.GetUserActiveSession(user.UserId)
		if sessionErr != nil {
			return nil, NewSessionError("getting user session", sessionErr.Error())
		}

		switch user.Status {
		case userModels.Deleted:
			return nil, NewUserDeletedError(user.Email)
		case userModels.Banned:
			return nil, NewUserBannedError(user.Email, "Banned due to violation of terms of service")
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
			_, err = s.sessionService.CreateSession(user.UserId, token.AccessToken, token.RefreshToken, token.AccessTokenExpiresIn.Time(), deviceInfo, locationInfo)
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

	// 1. Find user by email
	user, err := s.userService.FindUserByEmail(forgotPasswordRequest.Email)
	if err != nil {
		return err
	}
	if user == nil {
		return NewUserNotFoundError(forgotPasswordRequest.Email)
	}

	// 2. Generate secure token (32 bytes, hex encoded = 64 characters)
	token := generateSecureToken(32)
	expiry := time.Now().Add(1 * time.Hour)

	// 3. Save token to user model
	err = s.userService.UpdatePasswordResetToken(user.UserId, token, expiry)
	if err != nil {
		s.logger.Error("Failed to update password reset token: %v", err)
		return NewTokenError("generating password reset token", err.Error())
	}

	// 4. Send email via EmailService
	resetUrl := fmt.Sprintf("%s/reset-password?token=%s", s.env.AppFrontendURL, token)
	emailErr := s.emailService.SendPasswordReset(user.Email, token, resetUrl)
	if emailErr != nil {
		s.logger.Error("Failed to send password reset email: %v", emailErr)
		return NewEmailSendError("password reset", emailErr)
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

func (s *authService) VerifyEmail(token string) (*userModels.User, network.ApiError) {
	s.logger.Info("Verifying email with token")

	// 1. Find user by token
	user, err := s.userService.FindUserByEmailVerificationToken(token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, NewInvalidTokenError("email verification")
	}

	// 2. Check token expiry
	if user.EmailVerificationExpiry != nil && user.EmailVerificationExpiry.Time().Before(time.Now()) {
		s.logger.Error("Email verification token expired for user: %s", user.Email)
		return nil, NewExpiredTokenError("email verification")
	}

	// 3. Check if already verified
	if user.VerifiedEmail {
		s.logger.Info("Email already verified for user: %s", user.Email)
		return nil, NewEmailAlreadyVerifiedError(user.Email)
	}

	// 4. Mark email as verified and clear token
	err = s.userService.MarkEmailAsVerified(user.UserId)
	if err != nil {
		s.logger.Error("Failed to mark email as verified: %v", err)
		return nil, err
	}

	// 5. Get updated user
	updatedUser, err := s.userService.FindUserById(user.UserId)
	if err != nil {
		return nil, err
	}

	s.logger.Success("Email verified successfully for user: %s", user.Email)
	return updatedUser, nil
}

func (s *authService) ResetPassword(token string, newPassword string) network.ApiError {
	s.logger.Info("Resetting password with token")

	// 1. Find user by reset token
	user, err := s.userService.FindUserByPasswordResetToken(token)
	if err != nil {
		return err
	}
	if user == nil {
		return NewInvalidTokenError("password reset")
	}

	// 2. Check token expiry
	if user.PasswordResetExpiry != nil && user.PasswordResetExpiry.Time().Before(time.Now()) {
		s.logger.Error("Password reset token expired for user: %s", user.Email)
		return NewExpiredTokenError("password reset")
	}

	// 3. Hash new password
	hashedPassword, hashErr := utils.HashPassword(newPassword)
	if hashErr != nil {
		s.logger.Error("Failed to hash password: %v", hashErr)
		return NewUserError("hashing password", hashErr.Error())
	}

	// 4. Update password and clear reset token
	err = s.userService.UpdatePasswordWithResetToken(user.UserId, hashedPassword)
	if err != nil {
		s.logger.Error("Failed to update password: %v", err)
		return err
	}

	// 5. Invalidate all user sessions (security measure)
	sessions, sessionErr := s.sessionService.GetActiveSessionsByUserID(user.UserId)
	if sessionErr != nil {
		s.logger.Error("Failed to get active sessions: %v", sessionErr)
		// Don't fail the request, just log the error
	} else {
		for _, session := range sessions {
			invalidateErr := s.sessionService.InvalidateSession(session.SessionID)
			if invalidateErr != nil {
				s.logger.Error("Failed to invalidate session %s: %v", session.SessionID, invalidateErr)
			}
		}
	}

	s.logger.Success("Password reset successfully for user: %s", user.Email)
	return nil
}

// Helper function to generate cryptographically secure random token
func generateSecureToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
