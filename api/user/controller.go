package user

import (
	"sync-backend/api/common/location"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type userController struct {
	logger utils.AppLogger
	network.BaseController
	common.ContextPayload
	authProvider    network.AuthenticationProvider
	userService     UserService
	locationService location.LocationService
}

func NewUserController(
	authProvider network.AuthenticationProvider,
	userService UserService,
	locationService location.LocationService,
) network.Controller {
	return &userController{
		logger:          utils.NewServiceLogger("UserController"),
		BaseController:  network.NewBaseController("/api/v1/user", nil),
		ContextPayload:  common.NewContextPayload(),
		authProvider:    authProvider,
		userService:     userService,
		locationService: locationService,
	}
}

func (c *userController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting user routes")
	group.GET("/me", c.authProvider.Middleware(), c.GetMe)
}

func (c *userController) GetMe(ctx *gin.Context) {

	userId := c.ContextPayload.MustGetUserId(ctx)
	user, err := c.userService.FindUserById(*userId)

	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if user == nil {
		c.Send(ctx).NotFoundError("User not found", nil)
		return
	}

	c.Send(ctx).SuccessDataResponse("Profile fetched successfully", user)
}
