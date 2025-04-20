package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/api/user"
	"sync-backend/arch/common"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type authController struct {
	logger utils.AppLogger
	network.BaseController
	common.ContextPayload
	authService AuthService
	userService user.UserService
}

func NewAuthController(
	logger utils.AppLogger,
	authService AuthService,
	userService user.UserService,
	authProvider network.AuthenticationProvider,
) network.Controller {
	return &authController{
		logger:         logger,
		BaseController: network.NewBaseController("/api/v1/auth", authProvider),
		ContextPayload: common.NewContextPayload(),
		authService:    authService,
		userService:    userService,
	}
}

func (c *authController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting auth routes")
	group.POST("/signup", c.SignUp)
	group.POST("/login", c.Login)
}

func (c *authController) SignUp(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewSignUpRequest())
	if err != nil {
		return
	}
	exists, err := c.userService.FindUserByEmail(body.Email)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		c.Send(ctx).MixedError(err)
		return
	}
	if exists != nil {
		c.Send(ctx).ConflictError("User with this email already exists", nil)
		return
	}

	data, err := c.authService.SignUp(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User created successfully", data)
}

func (c *authController) Login(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewLoginRequest())
	if err != nil {
		return
	}
	data, err := c.authService.Login(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User logged in successfully", data)
}
