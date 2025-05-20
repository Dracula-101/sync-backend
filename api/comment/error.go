package comment

import (
	"fmt"
	"sync-backend/arch/network"
)

const (
	ERR_COMMENT_NOT_FOUND   = "ERR_COMMENT_NOT_FOUND"
	ERR_POST_NOT_FOUND      = "ERR_POST_NOT_FOUND"
	ERR_COMMUNITY_NOT_FOUND = "ERR_COMMUNITY_NOT_FOUND"
	ERR_DB                  = "ERR_DB"
	ERR_FORBIDDEN           = "ERR_FORBIDDEN"
)

func NewCommentNotFoundError(commentId string) network.ApiError {
	return network.NewNotFoundError(
		fmt.Sprintf("Comment with ID '%s' not found. It may have been deleted or never existed. [Context: commentId=%s]", commentId, commentId),
		nil,
	)
}

func NewPostNotFoundError(postId string) network.ApiError {
	return network.NewNotFoundError(
		fmt.Sprintf("Post with ID '%s' not found. [Context: postId=%s]", postId, postId),
		nil,
	)
}

func NewCommunityNotFoundError(communityId string) network.ApiError {
	return network.NewNotFoundError(
		fmt.Sprintf("Community with ID '%s' not found. [Context: communityId=%s]", communityId, communityId),
		nil,
	)
}

func NewForbiddenError(action, userId, commentId string) network.ApiError {
	return network.NewForbiddenError(
		fmt.Sprintf("User '%s' is not authorized to %s comment '%s'. [Context: userId=%s, commentId=%s]", userId, action, commentId, userId, commentId),
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
