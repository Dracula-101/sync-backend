package post

import (
	"fmt"
	"sync-backend/arch/network"
)

const (
	ERR_POST_NOT_FOUND = "ERR_POST_NOT_FOUND"
	ERR_DB             = "ERR_DB"
	ERR_FORBIDDEN      = "ERR_FORBIDDEN"
	ERR_MEDIA          = "ERR_MEDIA"
)

func NewPostNotFoundError(postId string) network.ApiError {
	return network.NewNotFoundError(
		fmt.Sprintf("Post with ID '%s' not found. It may have been deleted or never existed. [Context: postId=%s]", postId, postId),
		nil,
	)
}

func NewDBError(action, extra string) network.ApiError {
	return network.NewInternalServerError(
		fmt.Sprintf("Database error occurred during %s. Details: %s", action, extra),
		ERR_DB,
		nil,
	)
}

func NewForbiddenError(action, userId, postId string) network.ApiError {
	return network.NewForbiddenError(
		fmt.Sprintf("User '%s' is not authorized to %s post '%s'. [Context: userId=%s, postId=%s]", userId, action, postId, userId, postId),
		nil,
	)
}

func NewMediaError(action, extra string) network.ApiError {
	return network.NewInternalServerError(
		fmt.Sprintf("Media error occurred during %s. Details: %s", action, extra),
		ERR_MEDIA,
		nil,
	)
}
