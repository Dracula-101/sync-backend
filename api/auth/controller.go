package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/arch/network"
	"sync-backend/common"

	"github.com/gin-gonic/gin"
)

type authController struct {
	network.BaseController
	common.ContextPayload
	service AuthService
}

func NewAuthController(
	service AuthService,
) network.Controller {
	return &authController{
		BaseController: network.NewBaseController("/api/v1/auth"),
		ContextPayload: common.NewContextPayload(),
		service:        service,
	}
}

func (c *authController) MountRoutes(group *gin.RouterGroup) {
	group.POST("/ping", c.Ping)
	group.POST("/signup", c.SignUp)
}

func (c *authController) Ping(ctx *gin.Context) {
	c.Send(ctx).SuccessMsgResponse("pong")
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
