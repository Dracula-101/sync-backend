package community

import (
	"fmt"
	"sync-backend/arch/network"
)

const (
	ERR_COMMUNITY_NOT_FOUND = "ERR_COMMUNITY_NOT_FOUND"
	ERR_DB                  = "ERR_DB"
	ERR_FORBIDDEN           = "ERR_FORBIDDEN"
	ERR_DUPLICATE           = "ERR_DUPLICATE"
)

func NewCommunityNotFoundError(communityId string) network.ApiError {
	return network.NewNotFoundError(
		"Community Not Found",
		fmt.Sprintf("Community with ID '%s' not found. It may have been deleted or never existed. [Context: communityId=%s]", communityId, communityId),
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

func NewForbiddenError(action, userId, communityId string) network.ApiError {
	return network.NewForbiddenError(
		"Forbidden",
		fmt.Sprintf("User '%s' is not authorized to %s community '%s'. [Context: userId=%s, communityId=%s]", userId, action, communityId, userId, communityId),
		nil,
	)
}

func NewDuplicateCommunityError(slug string) network.ApiError {
	return network.NewConflictError(
		"Duplicate Community",
		fmt.Sprintf("A community with the slug '%s' already exists. Please choose a different name. [Context: slug=%s]", slug, slug),
		nil,
	)
}

func NewNotAuthorizedError(action, userId, communityId string) network.ApiError {
	return network.NewForbiddenError(
		"Not Authorized",
		fmt.Sprintf("User '%s' is not authorized to %s community '%s'. [Context: userId=%s, communityId=%s]", userId, action, communityId, userId, communityId),
		nil,
	)
}
