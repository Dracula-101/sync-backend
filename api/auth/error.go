package auth

import (
	"fmt"
	"sync-backend/arch/network"
)

const (
	ERR_USER                   = "ERR_USER"
	ERR_BAN_USER               = "ERR_BAN_USER"
	ERR_TOKEN                  = "ERR_TOKEN"
	ERR_SESSION                = "ERR_SESSION"
	ERR_USER_NOT_FOUND         = "ERR_USER_NOT_FOUND"
	ERR_SESSION_NOT_FOUND      = "ERR_SESSION_NOT_FOUND"
	ERR_SESSION_EXPIRED        = "ERR_SESSION_EXPIRED"
	ERR_SESSION_INVALID        = "ERR_SESSION_INVALID"
	ERR_INVALID_TOKEN          = "ERR_INVALID_TOKEN"
	ERR_EXPIRED_TOKEN          = "ERR_EXPIRED_TOKEN"
	ERR_EMAIL_ALREADY_VERIFIED = "ERR_EMAIL_ALREADY_VERIFIED"
	ERR_EMAIL_SEND_FAILED      = "ERR_EMAIL_SEND_FAILED"
)

// User not found error
func NewUserNotFoundError(email string) network.ApiError {
	return network.NewNotFoundError(
		"User with email not found",
		fmt.Sprintf("This may indicate the user never registered or was deleted. [Context: email=%s]", email),
		nil,
	)
}

// User banned error
func NewUserBannedError(email string, banReason string) network.ApiError {
	return network.NewBadRequestError(
		fmt.Sprintf("User is banned. Reason - %s", banReason),
		fmt.Sprintf("This may indicate the user is banned and can't use the platform. [Context: email=%s]", email),
		nil,
	)
}

// User deleted error
func NewUserDeletedError(email string) network.ApiError {
	return network.NewBadRequestError(
		"User is deleted or unavailable.",
		fmt.Sprintf("This account is deleted or marked for deletion, follow the instruction to recover account. [Context: email=%s]", email),
		nil,
	)
}

// Invalid password error
func NewInvalidPasswordError(email string) network.ApiError {
	return network.NewUnauthorizedError(
		"Entered password is incorrect for user",
		fmt.Sprintf("This may indicate the user entered an incorrect password. [Context: email=%s]", email),
		nil,
	)
}

// User has not set password error
func NewUserNoPasswordError(email string) network.ApiError {
	return network.NewBadRequestError(
		"User has not set password",
		fmt.Sprintf("This may indicate the user has not set a password or is using a third-party login method. [Context: email=%s]", email),
		nil,
	)
}

// User already exists error (by email)
func NewUserExistsByEmailError(email string) network.ApiError {
	return network.NewConflictError(
		"User with this email already exists",
		fmt.Sprintf("This may indicate the user is already registered. [Context: email=%s]", email),
		nil,
	)
}

// User already exists error (by username)
func NewUserExistsByUsernameError(username string) network.ApiError {
	return network.NewConflictError(
		"User with this username already exists",
		fmt.Sprintf("This may indicate the user is already registered. [Context: username=%s]", username),
		nil,
	)
}

// Session not found error
func NewSessionNotFoundError(userId string) network.ApiError {
	return network.NewInternalServerError(
		"Session not found for user",
		fmt.Sprintf("This may indicate the user has not logged in or the session has expired. [Context: userId=%s]", userId),
		ERR_SESSION_NOT_FOUND,
		nil,
	)
}

// Session invalid error
func NewSessionInvalidError(sessionId string) network.ApiError {
	return network.NewInternalServerError(
		"Session is invalid",
		fmt.Sprintf("This may indicate the session is invalid or has been tampered with. [Context: sessionId=%s]", sessionId),
		ERR_SESSION_INVALID,
		nil,
	)
}

// Session expired error
func NewSessionExpiredError(sessionId string) network.ApiError {
	return network.NewInternalServerError(
		"Session has expired",
		fmt.Sprintf("This may indicate the session has expired and needs to be refreshed. [Context: sessionId=%s]", sessionId),
		ERR_SESSION_EXPIRED,
		nil,
	)
}

// Token error
func NewTokenError(context string, extra string) network.ApiError {
	return network.NewInternalServerError(
		"Token error occurred",
		fmt.Sprintf("This may indicate the token is invalid or has expired. [Context: %s] [Extra: %s]", context, extra),
		ERR_TOKEN,
		nil,
	)
}

// General user error
func NewUserError(context string, extra string) network.ApiError {
	return network.NewInternalServerError(
		"User error occurred",
		fmt.Sprintf("This may indicate an issue with the user account. [Context: %s] [Extra: %s]", context, extra),
		ERR_USER,
		nil,
	)
}

// General session error
func NewSessionError(context string, extra string) network.ApiError {
	return network.NewInternalServerError(
		"Session error occurred",
		fmt.Sprintf("This may indicate an issue with the session. [Context: %s] [Extra: %s]", context, extra),
		ERR_SESSION,
		nil,
	)
}

// Invalid token error (for email verification or password reset)
func NewInvalidTokenError(tokenType string) network.ApiError {
	return network.NewBadRequestError(
		"Invalid token",
		fmt.Sprintf("The %s token is invalid or has already been used", tokenType),
		nil,
	)
}

// Expired token error (for email verification or password reset)
func NewExpiredTokenError(tokenType string) network.ApiError {
	return network.NewBadRequestError(
		"Token expired",
		fmt.Sprintf("The %s token has expired. Please request a new one", tokenType),
		nil,
	)
}

// Email already verified error
func NewEmailAlreadyVerifiedError(email string) network.ApiError {
	return network.NewBadRequestError(
		"Email already verified",
		fmt.Sprintf("The email %s has already been verified", email),
		nil,
	)
}

// Email send failed error
func NewEmailSendError(emailType string, err error) network.ApiError {
	return network.NewInternalServerError(
		"Email send failed",
		fmt.Sprintf("Failed to send %s email. Please try again later", emailType),
		ERR_EMAIL_SEND_FAILED,
		err,
	)
}
