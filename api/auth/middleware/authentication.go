package middleware

import (
	"sync-backend/api/auth"
	"sync-backend/api/user"
	"sync-backend/arch/network"
	"sync-backend/common"

	"github.com/gin-gonic/gin"
)

type authenticationProvider struct {
	network.ResponseSender
	common.ContextPayload
	authService auth.AuthService
	userService user.UserService
}

func NewAuthenticationProvider(
	authService auth.AuthService,
	userService user.UserService,
) *authenticationProvider {
	return &authenticationProvider{
		ResponseSender: network.NewResponseSender(),
		authService:    authService,
		userService:    userService,
	}
}

func (p *authenticationProvider) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(network.AuthorizationHeader)
		if tokenString == "" {
			p.Send(ctx).UnauthorizedError("permission denied: no token provided", nil)
			return
		}

		token, claims, err := p.authService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			p.Send(ctx).UnauthorizedError("permission denied: invalid token", nil)
			return
		}

		userId := claims["user_id"].(string)
		user, err := p.userService.GetUserById(userId)
		if err != nil {
			p.Send(ctx).UnauthorizedError("permission denied: user not found", nil)
			return
		}

		p.SetUser(ctx, user)
		ctx.Next()
	}
}
