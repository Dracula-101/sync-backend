package middleware

import (
	"strings"
	"sync-backend/api/token"
	"sync-backend/api/user"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type authenticationProvider struct {
	network.ResponseSender
	common.ContextPayload
	logger       utils.AppLogger
	tokenService token.TokenService
	userService  user.UserService
}

func NewAuthenticationProvider(
	logger utils.AppLogger,
	tokenService token.TokenService,
	userService user.UserService,
) *authenticationProvider {
	return &authenticationProvider{
		ResponseSender: network.NewResponseSender(),
		ContextPayload: common.NewContextPayload(),
		logger:         logger,
		tokenService:   tokenService,
		userService:    userService,
	}
}

func (p *authenticationProvider) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(network.AuthorizationHeader)
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			p.Send(ctx).UnauthorizedError("Invalid or missing Authorization header", nil)
			ctx.Abort()
			return
		}
		tokenString := strings.Split(authHeader, " ")[1]

		token, claims, err := p.tokenService.ValidateToken(tokenString)
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

		p.SetUserId(ctx, claims.UserID)
		ctx.Next()
	}
}
