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
		"Comment Not Found",
		fmt.Sprintf("It seems the comment with ID '%s' does not exist. The comment may have been deleted or the comment ID is incorrect. [Context: commentId=%s]", commentId, commentId),
		nil,
	)
}

func NewPostNotFoundError(postId string) network.ApiError {
	return network.NewNotFoundError(
		"Post Not Found",
		fmt.Sprintf("It seems the post with ID '%s' does not exist. The post may have been deleted or the post ID is incorrect. [Context: postId=%s]", postId, postId),
		nil,
	)
}

func NewCommunityNotFoundError(communityId string) network.ApiError {
	return network.NewNotFoundError(
		"Community Not Found",
		fmt.Sprintf("It seems the community with ID '%s' does not exist. The community may have been deleted or the community ID is incorrect. [Context: communityId=%s]", communityId, communityId),
		nil,
	)
}

func NewForbiddenError(action, userId, commentId string) network.ApiError {
	return network.NewForbiddenError(
		"Forbidden",
		fmt.Sprintf("You do not have permission to perform this action. [Context: action=%s, userId=%s, commentId=%s]", action, userId, commentId),
		nil,
	)
}

func NewDBError(action, extra string) network.ApiError {
	return network.NewInternalServerError(
		"Database Error",
		fmt.Sprintf("An error occurred while performing the action '%s'. Please try again later. [Context: action=%s, extra=%s]", action, action, extra),
		ERR_DB,
		nil,
	)
}
