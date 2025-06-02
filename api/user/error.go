package user

import (
	"fmt"
	"sync-backend/arch/network"
)

const (
	ERR_USER_NOT_FOUND       = "ERR_USER_NOT_FOUND"
	ERR_USER_EXISTS_EMAIL    = "ERR_USER_EXISTS_EMAIL"
	ERR_USER_EXISTS_USERNAME = "ERR_USER_EXISTS_USERNAME"
	ERR_USER_BANNED          = "ERR_USER_BANNED"
	ERR_USER_DELETED         = "ERR_USER_DELETED"
	ERR_USER_INACTIVE        = "ERR_USER_INACTIVE"
	ERR_DB                   = "ERR_DB"
	ERR_FORBIDDEN            = "ERR_FORBIDDEN"
)

func NewUserNotFoundError(userId string) network.ApiError {
	return network.NewNotFoundError(
		"User Not Found",
		fmt.Sprintf("User with ID '%s' not found. It may have been deleted, banned, or never existed. [Context: userId=%s]", userId, userId),
		nil,
	)
}

func NewUserNotFoundByEmailError(email string) network.ApiError {
	return network.NewNotFoundError(
		"User with email not found",
		fmt.Sprintf("User with email '%s' not found. It may have been deleted, banned, or never existed. [Context: email=%s]", email, email),
		nil,
	)
}

func NewUserNotFoundByUsernameError(username string) network.ApiError {
	return network.NewNotFoundError(
		"User with username not found",
		fmt.Sprintf("User with username '%s' not found. It may have been deleted, banned, or never existed. [Context: username=%s]", username, username),
		nil,
	)
}

func NewUserBannedError(userId string, banReason string) network.ApiError {
	return network.NewForbiddenError(
		"User Banned. Reason: "+banReason,
		fmt.Sprintf("User '%s' is banned and cannot perform this action. [Context: userId=%s]", userId, userId),
		nil,
	)
}

func NewUserMarkedForDeletionError(userId string) network.ApiError {
	return network.NewForbiddenError(
		"User Marked for Deletion",
		fmt.Sprintf("User '%s' is marked for deletion and cannot perform this action. [Context: userId=%s]", userId, userId),
		nil,
	)
}

func NewUserDeletedError(userId string) network.ApiError {
	return network.NewForbiddenError(
		"User Deleted",
		fmt.Sprintf("User '%s' is deleted and cannot perform this action. [Context: userId=%s]", userId, userId),
		nil,
	)
}

func NewUserInactiveError(userId string) network.ApiError {
	return network.NewForbiddenError(
		"User Inactive",
		fmt.Sprintf("User '%s' is inactive and cannot perform this action. [Context: userId=%s]", userId, userId),
		nil,
	)
}

func NewUserExistsByEmailError(email string) network.ApiError {
	return network.NewConflictError(
		"A user with this email already exists",
		fmt.Sprintf("A user with email '%s' already exists. Registration or update cannot proceed. [Context: email=%s]", email, email),
		nil,
	)
}

func NewUserExistsByUsernameError(username string) network.ApiError {
	return network.NewConflictError(
		"A user with this username already exists",
		fmt.Sprintf("A user with username '%s' already exists. Registration or update cannot proceed. [Context: username=%s]", username, username),
		nil,
	)
}

func NewWrongOldPasswordError(userId string) network.ApiError {
	return network.NewUnauthorizedError(
		"Incorrect Old Password",
		fmt.Sprintf("The old password provided for user '%s' is incorrect. Please try again. [Context: userId=%s]", userId, userId),
		nil,
	)
}

func NewWrongPasswordError(userId string) network.ApiError {
	return network.NewUnauthorizedError(
		"Incorrect Password",
		fmt.Sprintf("The password provided for user '%s' is incorrect. Please try again. [Context: userId=%s]", userId, userId),
		nil,
	)
}

func NewForbiddenUserActionError(action, userId, targetId string) network.ApiError {
	return network.NewForbiddenError(
		"Forbidden",
		fmt.Sprintf("User '%s' is not authorized to %s user '%s'. [Context: userId=%s, targetId=%s]", userId, action, targetId, userId, targetId),
		nil,
	)
}

func NewDBError(action, extra string) network.ApiError {
	return network.NewInternalServerError(
		"Database Error",
		fmt.Sprintf("Database error occurred during %s. Details: %s", action, extra),
		ERR_DB,
		nil,
	)
}

func NewSelfActionError(action string) network.ApiError {
	return network.NewBadRequestError(
		"Self Action Not Allowed",
		fmt.Sprintf("You cannot %s yourself. This action is not allowed.", action),
		nil,
	)
}
