package middleware

import (
	"sync-backend/api/token"
	"sync-backend/api/user"
	"sync-backend/arch/common"
	"sync-backend/arch/network"

	"github.com/gin-gonic/gin"
)

type authenticationProvider struct {
	network.ResponseSender
	common.ContextPayload
	tokenService token.TokenService
	userService  user.UserService
}

func NewAuthenticationProvider(
	tokenService token.TokenService,
	userService user.UserService,
) *authenticationProvider {
	return &authenticationProvider{
		ResponseSender: network.NewResponseSender(),
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
		tokenString := authHeader[7:]

		token, claims, err := p.tokenService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			p.Send(ctx).UnauthorizedError("permission denied: invalid token", nil)
			return
		}

		p.SetUserId(ctx, claims.UserID)
		ctx.Next()
	}
}
