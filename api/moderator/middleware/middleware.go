package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync-backend/api/moderator"
	"sync-backend/api/moderator/model"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/arch/redis"
	"sync-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type ModeratorMiddleware interface {
	RequiresModerator(communityIdParam string) gin.HandlerFunc
	RequiresPermission(communityIdParam string, permission model.ModeratorPermission) gin.HandlerFunc
	RequiresAdmin(communityIdParam string) gin.HandlerFunc
	CheckUserNotBanned(communityIdParam string) gin.HandlerFunc
}

type moderatorMiddleware struct {
	network.ResponseSender
	common.ContextPayload
	logger           utils.AppLogger
	moderatorService moderator.ModeratorService
	contextPayload   common.ContextPayload
	store            redis.Store
}

// NewModeratorMiddleware creates a new moderator auth middleware
func NewModeratorMiddleware(moderatorService moderator.ModeratorService, store redis.Store) ModeratorMiddleware {
	return &moderatorMiddleware{
		ResponseSender:   network.NewResponseSender(),
		ContextPayload:   common.NewContextPayload(),
		logger:           utils.NewServiceLogger("ModeratorAuthMiddleware"),
		contextPayload:   common.NewContextPayload(),
		moderatorService: moderatorService,
		store:            store,
	}
}

// RequiresModerator middleware ensures the user is a moderator for the community
func (m *moderatorMiddleware) RequiresModerator(communityIdParam string) gin.HandlerFunc {
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

		// Check cache first
		cacheKey := fmt.Sprintf("moderator:%s:%s", *userId, communityId)
		var isModerator bool
		result, err := m.store.GetInstance().Get(context.Background(), cacheKey).Result()
		if err == nil {
			// Successfully got from cache
			if err := json.Unmarshal([]byte(result), &isModerator); err == nil {
				if !isModerator {
					m.Send(ctx).ForbiddenError(
						"Permission denied",
						"You must be a moderator to access this resource",
						errors.New("not a moderator"),
					)
					return
				}
				ctx.Next()
				return
			}
		}

		// Cache miss, query the database
		isModerator, apiErr := m.moderatorService.IsModeratorOrHigher(*userId, communityId)
		if apiErr != nil {
			m.Send(ctx).MixedError(apiErr)
			return
		}

		// Cache result (5 minutes TTL)
		moderatorJSON, _ := json.Marshal(isModerator)
		m.store.GetInstance().Set(context.Background(), cacheKey, moderatorJSON, time.Second*300)

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
func (m *moderatorMiddleware) RequiresPermission(communityIdParam string, permission model.ModeratorPermission) gin.HandlerFunc {
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

		// Check cache first
		cacheKey := fmt.Sprintf("permission:%s:%s:%s", *userId, communityId, permission)
		var hasPermission bool
		result, err := m.store.GetInstance().Get(context.Background(), cacheKey).Result()
		if err == nil {
			// Successfully got from cache
			if err := json.Unmarshal([]byte(result), &hasPermission); err == nil {
				if !hasPermission {
					m.Send(ctx).ForbiddenError(
						"Permission denied",
						fmt.Sprintf("You must have the '%s' permission to access this resource", permission),
						errors.New("insufficient permissions"),
					)
					return
				}
				ctx.Next()
				return
			}
		}

		// Cache miss, query the database
		hasPermission, apiErr := m.moderatorService.HasModeratorPermission(*userId, communityId, permission)
		if apiErr != nil {
			m.Send(ctx).MixedError(apiErr)
			return
		}

		// Cache result (5 minutes TTL)
		permissionJSON, _ := json.Marshal(hasPermission)
		m.store.GetInstance().Set(context.Background(), cacheKey, permissionJSON, time.Second*300)

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
func (m *moderatorMiddleware) RequiresAdmin(communityIdParam string) gin.HandlerFunc {
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

		// Check cache first
		cacheKey := fmt.Sprintf("admin:%s:%s", *userId, communityId)
		var isAdmin bool
		result, err := m.store.GetInstance().Get(context.Background(), cacheKey).Result()
		if err == nil {
			// Successfully got from cache
			if err := json.Unmarshal([]byte(result), &isAdmin); err == nil {
				if !isAdmin {
					m.Send(ctx).ForbiddenError(
						"Permission denied",
						"You must be an admin to access this resource",
						errors.New("not an admin or owner"),
					)
					return
				}
				ctx.Next()
				return
			}
		}

		// Cache miss, query the database
		isAdmin, apiErr := m.moderatorService.IsAdmin(*userId, communityId)
		if apiErr != nil {
			m.Send(ctx).MixedError(apiErr)
			return
		}

		// Cache result (5 minutes TTL)
		adminJSON, _ := json.Marshal(isAdmin)
		m.store.GetInstance().Set(context.Background(), cacheKey, adminJSON, time.Second*300)

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

// CheckUserNotBanned middleware ensures the user is not banned from the community
func (m *moderatorMiddleware) CheckUserNotBanned(communityIdParam string) gin.HandlerFunc {
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

		// Check cache first
		cacheKey := fmt.Sprintf("banned:%s:%s", *userId, communityId)
		var banData struct {
			IsBanned bool
			BanInfo  *model.BanInfo
		}

		result, err := m.store.GetInstance().Get(context.Background(), cacheKey).Result()
		if err == nil {
			// Successfully got from cache
			if err := json.Unmarshal([]byte(result), &banData); err == nil {
				if banData.IsBanned {
					var banMessage string
					if banData.BanInfo.IsPermanent {
						banMessage = fmt.Sprintf("You are permanently banned from this community. Reason: %s", banData.BanInfo.Reason)
					} else {
						banMessage = fmt.Sprintf("You are banned from this community until %s. Reason: %s",
							banData.BanInfo.ExpiresAt.Format("Jan 2, 2006"), banData.BanInfo.Reason)
					}

					m.Send(ctx).ForbiddenError(
						"Access denied",
						banMessage,
						errors.New("user is banned from the community"),
					)
					return
				}
				ctx.Next()
				return
			}
		}

		// Cache miss, query the database
		isBanned, banInfo, apiErr := m.moderatorService.IsUserBanned(*userId, communityId)
		if apiErr != nil {
			m.Send(ctx).MixedError(apiErr)
			return
		}

		// Cache result (1 minute TTL for bans since they might change frequently)
		banData = struct {
			IsBanned bool
			BanInfo  *model.BanInfo
		}{
			IsBanned: isBanned,
			BanInfo:  banInfo,
		}
		banDataJSON, _ := json.Marshal(banData)
		m.store.GetInstance().Set(context.Background(), cacheKey, banDataJSON, time.Second*60)

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
