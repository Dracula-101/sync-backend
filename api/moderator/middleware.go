package moderator

import (
	"errors"
	"fmt"
	"sync-backend/api/moderator/model"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type ModeratorAuthMiddleware interface {
	RequiresModerator(communityIdParam string) gin.HandlerFunc
	RequiresPermission(communityIdParam string, permission model.ModeratorPermission) gin.HandlerFunc
	RequiresAdmin(communityIdParam string) gin.HandlerFunc
}

type moderatorAuthMiddleware struct {
	network.ResponseSender
	common.ContextPayload
	logger           utils.AppLogger
	moderatorService ModeratorService
	contextPayload   common.ContextPayload
}

// NewModeratorAuthMiddleware creates a new moderator auth middleware
func NewModeratorAuthMiddleware(moderatorService ModeratorService) ModeratorAuthMiddleware {
	return &moderatorAuthMiddleware{
		ResponseSender:   network.NewResponseSender(),
		ContextPayload:   common.NewContextPayload(),
		logger:           utils.NewServiceLogger("ModeratorAuthMiddleware"),
		moderatorService: moderatorService,
		contextPayload:   common.NewContextPayload(),
	}
}

// RequiresModerator middleware ensures the user is a moderator for the community
func (m *moderatorAuthMiddleware) RequiresModerator(communityIdParam string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := m.contextPayload.MustGetUserId(ctx)
		if userId == nil {
			m.Send(ctx).UnauthorizedError(
				"Unauthorized",
				"You must be logged in to access this resource",
				nil,
			)
			return
		}

		communityId := ctx.Param(communityIdParam)
		if communityId == "" {
			m.Send(ctx).BadRequestError(
				"Invalid request",
				fmt.Sprintf("Missing community ID parameter: %s", communityIdParam),
				nil,
			)
			return
		}

		isModerator, apiErr := m.moderatorService.IsModeratorOrHigher(*userId, communityId)
		if apiErr != nil {
			m.Send(ctx).MixedError(apiErr)
			return
		}

		if !isModerator {
			m.Send(ctx).ForbiddenError(
				"Permission denied",
				"You must be a moderator to access this resource",
				errors.New("not a moderator"),
			)
			return
		}

		ctx.Next()
	}
}

// RequiresPermission middleware ensures the user has the specific permission for the community
func (m *moderatorAuthMiddleware) RequiresPermission(communityIdParam string, permission model.ModeratorPermission) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := m.contextPayload.MustGetUserId(ctx)
		if userId == nil {
			m.Send(ctx).UnauthorizedError(
				"Unauthorized",
				"You must be logged in to access this resource",
				nil,
			)
			return
		}

		communityId := ctx.Param(communityIdParam)
		if communityId == "" {
			m.Send(ctx).BadRequestError(
				"Invalid request",
				fmt.Sprintf("Missing community ID parameter: %s", communityIdParam),
				nil,
			)
			return
		}

		hasPermission, apiErr := m.moderatorService.HasModeratorPermission(*userId, communityId, permission)
		if apiErr != nil {
			m.Send(ctx).MixedError(apiErr)
			return
		}

		if !hasPermission {
			m.Send(ctx).ForbiddenError(
				"Permission denied",
				fmt.Sprintf("You must have the '%s' permission to access this resource", permission),
				errors.New("insufficient permissions"),
			)
			return
		}

		ctx.Next()
	}
}

// RequiresAdmin middleware ensures the user is an admin for the community
func (m *moderatorAuthMiddleware) RequiresAdmin(communityIdParam string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := m.contextPayload.MustGetUserId(ctx)
		if userId == nil {
			m.Send(ctx).UnauthorizedError(
				"Unauthorized",
				"You must be logged in to access this resource",
				nil,
			)
			return
		}

		communityId := ctx.Param(communityIdParam)
		if communityId == "" {
			m.Send(ctx).BadRequestError(
				"Invalid request",
				fmt.Sprintf("Missing community ID parameter: %s", communityIdParam),
				nil,
			)
			return
		}

		isAdmin, apiErr := m.moderatorService.IsAdmin(*userId, communityId)
		if apiErr != nil {
			m.Send(ctx).MixedError(apiErr)
			return
		}

		if !isAdmin {
			m.Send(ctx).ForbiddenError(
				"Permission denied",
				"You must be an admin to access this resource",
				errors.New("not an admin or owner"),
			)
			return
		}

		ctx.Next()
	}
}
