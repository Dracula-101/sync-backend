package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/arch/network"
	"sync-backend/arch/common"

	"github.com/gin-gonic/gin"
)

type authController struct {
	network.BaseController
	common.ContextPayload
	service AuthService
}

func NewAuthController(
	service AuthService,
	authProvider network.AuthenticationProvider,
) network.Controller {
	return &authController{
		BaseController: network.NewBaseController("/api/v1/auth", authProvider),
		ContextPayload: common.NewContextPayload(),
		service:        service,
	}
}

func (c *authController) MountRoutes(group *gin.RouterGroup) {
	group.POST("/signup", c.SignUp)
	group.POST("/login", c.Login)
}

func (c *authController) SignUp(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewSignUpRequest())
	if err != nil {
		c.Send(ctx).BadRequestError(err.Error(), err)
		return
	}
	data, err := c.service.SignUp(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User created successfully", data)
}

func (c *authController) Login(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewLoginRequest())
	if err != nil {
		c.Send(ctx).BadRequestError(err.Error(), err)
		return
	}
	data, err := c.service.Login(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User logged in successfully", data)
}
