package middleware

import (
	"errors"
	"fmt"
	moderatorService "sync-backend/api/moderator"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

// BanCheckMiddleware checks if a user is banned from a community
type BanCheckMiddleware interface {
	CheckUserNotBanned(communityIdParam string) gin.HandlerFunc
}

type banCheckMiddleware struct {
	network.ResponseSender
	common.ContextPayload
	logger           utils.AppLogger
	moderatorService moderatorService.ModeratorService
}

// NewBanCheckMiddleware creates a new ban check middleware
func NewBanCheckMiddleware(moderatorService moderatorService.ModeratorService) BanCheckMiddleware {
	return &banCheckMiddleware{
		ResponseSender:   network.NewResponseSender(),
		ContextPayload:   common.NewContextPayload(),
		logger:           utils.NewServiceLogger("BanCheckMiddleware"),
		moderatorService: moderatorService,
	}
}

// CheckUserNotBanned middleware ensures the user is not banned from the community
func (m *banCheckMiddleware) CheckUserNotBanned(communityIdParam string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := m.MustGetUserId(ctx)
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

		// Check for community ban
		isBanned, banInfo, apiErr := m.moderatorService.IsUserBanned(*userId, communityId)
		if apiErr != nil {
			m.Send(ctx).MixedError(apiErr)
			return
		}

		if isBanned {
			var banMessage string
			if banInfo.IsPermanent {
				banMessage = fmt.Sprintf("You are permanently banned from this community. Reason: %s", banInfo.Reason)
			} else {
				banMessage = fmt.Sprintf("You are banned from this community until %s. Reason: %s",
					banInfo.ExpiresAt.Format("Jan 2, 2006"), banInfo.Reason)
			}

			m.Send(ctx).ForbiddenError(
				"Access denied",
				banMessage,
				errors.New("user is banned from the community"),
			)
			return
		}

		ctx.Next()
	}
}
