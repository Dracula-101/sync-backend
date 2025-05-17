package middleware

import (
	"strings"
	"sync-backend/api/common/session"
	"sync-backend/api/common/token"
	"sync-backend/api/user"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/arch/redis"
	"sync-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type authenticationProvider struct {
	network.ResponseSender
	common.ContextPayload
	logger         utils.AppLogger
	tokenService   token.TokenService
	sessionService session.SessionService
	cacheStore     redis.Store
	userService    user.UserService
}

func NewAuthenticationProvider(
	tokenService token.TokenService,
	userService user.UserService,
	sessionService session.SessionService,
	cacheStore redis.Store,
) *authenticationProvider {
	return &authenticationProvider{
		ResponseSender: network.NewResponseSender(),
		ContextPayload: common.NewContextPayload(),
		logger:         utils.NewServiceLogger("AuthProvider"),
		tokenService:   tokenService,
		sessionService: sessionService,
		cacheStore:     cacheStore,
		userService:    userService,
	}
}

func (p *authenticationProvider) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		authHeader := ctx.GetHeader(network.AuthorizationHeader)
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			p.Send(ctx).UnauthorizedError("Invalid or missing Authorization header", nil)
			return
		}

		tokenSplit := strings.Split(authHeader, " ")
		if len(tokenSplit) != 2 {
			p.logger.Error("Invalid Authorization header format: %s", authHeader)
			p.Send(ctx).UnauthorizedError("Invalid Authorization header format", nil)
			return
		}

		tokenString := tokenSplit[len(tokenSplit)-1]

		token, claims, err := p.tokenService.ValidateToken(tokenString, true)
		if err != nil {
			p.logger.Error("Failed to validate token: %v %v", tokenString, err)
			p.Send(ctx).UnauthorizedError("Invalid or expired token", err)
			return
		}
		if !token.Valid {
			p.logger.Error("Token is not valid: %v", tokenString)
			p.Send(ctx).UnauthorizedError("Invalid token", nil)
			return
		}

		// check if cache has the user ID
		redisCmd := p.cacheStore.GetInstance().Get(ctx, claims.UserID)
		hasUserId, err := redisCmd.Bool()

		if err != nil && err.Error() != "redis: nil" {
			p.logger.Error("Failed to get user ID from cache: %v", err)
			p.Send(ctx).InternalServerError("Failed to get user ID from cache", network.CACHE_ERROR, err)
			return
		}

		if !hasUserId {
			session, err := p.sessionService.GetSessionByToken(tokenString)
			if err != nil {
				p.logger.Error("Failed to get session by token: %v", err)
				p.Send(ctx).UnauthorizedError("Invalid or expired session", err)
				return
			}
			if session == nil {
				p.logger.Error("Session not found for token: %s", tokenString)
				p.Send(ctx).UnauthorizedError("Invalid or expired session", nil)
				return
			}
			if session.IsRevoked {
				p.logger.Error("Session is revoked for token: %s", tokenString)
				p.Send(ctx).UnauthorizedError("Session is revoked", nil)
				return
			}

			//save to cache
			err = p.cacheStore.GetInstance().Set(ctx, claims.UserID, true, time.Hour*1).Err()
			if err != nil {
				p.logger.Error("Failed to set user ID in cache: %v", err)
				p.Send(ctx).InternalServerError("Failed to set user ID in cache", network.CACHE_ERROR, err)
				return
			}

			p.logger.Debug("Set user ID in cache: %s", claims.UserID)
		}

		p.SetUserId(ctx, claims.UserID)
		p.logger.Debug("User ID from token: %s", claims.UserID)
		ctx.Next()
	}
}
