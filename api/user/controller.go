package user

import (
	"sync-backend/api/common/location"
	"sync-backend/api/user/dto"
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
	// join community
	group.POST("/join/:communityId", c.authProvider.Middleware(), c.JoinCommunity)
	group.POST("/leave/:communityId", c.authProvider.Middleware(), c.LeaveCommunity)
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

func (c *userController) JoinCommunity(ctx *gin.Context) {
	params, err := network.ReqParams(ctx, dto.NewJoinCommunityRequest())

	if err != nil {
		return
	}

	communityId := params.CommunityId
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

	if user.IsInCommunity(communityId) {
		c.Send(ctx).BadRequestError("Already in community", nil)
		return
	}

	err = c.userService.JoinCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Joined community successfully")
}

func (c *userController) LeaveCommunity(ctx *gin.Context) {
	params, err := network.ReqParams(ctx, dto.NewLeaveCommunityRequest())

	if err != nil {
		return
	}

	communityId := params.CommunityId
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

	if !user.IsInCommunity(communityId) {
		c.Send(ctx).BadRequestError("Not in community", nil)
		return
	}

	err = c.userService.LeaveCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Left community successfully")
}