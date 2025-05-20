package auth

import (
	"fmt"
	"sync-backend/arch/network"
)

const (
	ERR_USER              = "ERR_USER"
	ERR_BAN_USER          = "ERR_BAN_USER"
	ERR_TOKEN             = "ERR_TOKEN"
	ERR_SESSION           = "ERR_SESSION"
	ERR_USER_NOT_FOUND    = "ERR_USER_NOT_FOUND"
	ERR_SESSION_NOT_FOUND = "ERR_SESSION_NOT_FOUND"
	ERR_SESSION_EXPIRED   = "ERR_SESSION_EXPIRED"
	ERR_SESSION_INVALID   = "ERR_SESSION_INVALID"
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
		fmt.Sprintf("User is banned. Reason - %s", email, banReason),
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
