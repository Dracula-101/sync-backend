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
		fmt.Sprintf("User with email '%s' not found. This may indicate the user never registered or was deleted. [Context: email=%s]", email, email),
		nil,
	)
}

// User banned error
func NewUserBannedError(email string) network.ApiError {
	return network.NewBadRequestError(
		fmt.Sprintf("User '%s' is banned. Access denied. [Context: email=%s]", email, email),
		nil,
	)
}

// User deleted error
func NewUserDeletedError(email string) network.ApiError {
	return network.NewBadRequestError(
		fmt.Sprintf("User '%s' is deleted or unavailable. [Context: email=%s]", email, email),
		nil,
	)
}

// Invalid password error
func NewInvalidPasswordError(email string) network.ApiError {
	return network.NewUnauthorizedError(
		fmt.Sprintf("Entered password is incorrect for user '%s'. [Context: email=%s]", email, email),
		nil,
	)
}

// User has not set password error
func NewUserNoPasswordError(email string) network.ApiError {
	return network.NewBadRequestError(
		fmt.Sprintf("User '%s' has not set password. [Context: email=%s]", email, email),
		nil,
	)
}

// User already exists error (by email)
func NewUserExistsByEmailError(email string) network.ApiError {
	return network.NewConflictError(
		fmt.Sprintf("User with this email '%s' already exists. Registration cannot proceed. [Context: email=%s]", email, email),
		nil,
	)
}

// User already exists error (by username)
func NewUserExistsByUsernameError(username string) network.ApiError {
	return network.NewConflictError(
		fmt.Sprintf("User with this username '%s' already exists. Registration cannot proceed. [Context: username=%s]", username, username),
		nil,
	)
}

// Session not found error
func NewSessionNotFoundError(userId string) network.ApiError {
	return network.NewInternalServerError(
		fmt.Sprintf("Session not found for user '%s'. This may indicate the user is not logged in or session expired. [Context: userId=%s]", userId, userId),
		ERR_SESSION_NOT_FOUND,
		nil,
	)
}

// Session invalid error
func NewSessionInvalidError(sessionId string) network.ApiError {
	return network.NewInternalServerError(
		fmt.Sprintf("Session '%s' is invalid or could not be invalidated. [Context: sessionId=%s]", sessionId, sessionId),
		ERR_SESSION_INVALID,
		nil,
	)
}

// Session expired error
func NewSessionExpiredError(sessionId string) network.ApiError {
	return network.NewInternalServerError(
		fmt.Sprintf("Session '%s' has expired. Please log in again. [Context: sessionId=%s]", sessionId, sessionId),
		ERR_SESSION_EXPIRED,
		nil,
	)
}

// Token error
func NewTokenError(context string, extra string) network.ApiError {
	return network.NewInternalServerError(
		fmt.Sprintf("Token error occurred during %s. Details: %s", context, extra),
		ERR_TOKEN,
		nil,
	)
}

// General user error
func NewUserError(context string, extra string) network.ApiError {
	return network.NewInternalServerError(
		fmt.Sprintf("User error occurred during %s. Details: %s", context, extra),
		ERR_USER,
		nil,
	)
}

// General session error
func NewSessionError(context string, extra string) network.ApiError {
	return network.NewInternalServerError(
		fmt.Sprintf("Session error occurred during %s. Details: %s", context, extra),
		ERR_SESSION,
		nil,
	)
}
