package user

import (
	"fmt"
	"sync-backend/api/common/location"
	"sync-backend/api/user/dto"
	"sync-backend/arch/common"
	coreMW "sync-backend/arch/middleware"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type userController struct {
	logger utils.AppLogger
	network.BaseController
	common.ContextPayload
	authProvider    network.AuthenticationProvider
	uploadProvider  coreMW.UploadProvider
	userService     UserService
	locationService location.LocationService
}

func NewUserController(
	authProvider network.AuthenticationProvider,
	uploadProvider coreMW.UploadProvider,
	userService UserService,
	locationService location.LocationService,
) network.Controller {
	return &userController{
		logger:          utils.NewServiceLogger("UserController"),
		BaseController:  network.NewBaseController("/user", nil),
		ContextPayload:  common.NewContextPayload(),
		authProvider:    authProvider,
		uploadProvider:  uploadProvider,
		userService:     userService,
		locationService: locationService,
	}
}

func (c *userController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting user routes")
	group.Use(c.authProvider.Middleware())
	group.GET("/me", c.GetMe)
	group.GET("/:userId", c.GetUserById)
	group.POST("/follow/:userId", c.FollowUser)

	group.POST("/unfollow/:userId", c.UnfollowUser)
	group.POST("/block/:userId", c.BlockUser)
	group.POST("/unblock/:userId", c.UnblockUser)
}

func (c *userController) GetMe(ctx *gin.Context) {

	userId := c.ContextPayload.MustGetUserId(ctx)
	user, err := c.userService.FindUserById(*userId)

	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if user == nil {
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", *userId),
			nil,
		)
		return
	}

	c.Send(ctx).SuccessDataResponse("Profile fetched successfully", user)
}

func (c *userController) GetUserById(ctx *gin.Context) {
	params, err := network.ReqParams(ctx, dto.NewGetUserRequest())
	if err != nil {
		return
	}
	userId := params.UserId
	user, err := c.userService.FindUserById(userId)

	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if user == nil {
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", userId),
			nil,
		)
		return
	}

	c.Send(ctx).SuccessDataResponse("Profile fetched successfully", user)
}

func (c *userController) FollowUser(ctx *gin.Context) {
	followUserId := ctx.Param("userId")
	userId := c.ContextPayload.MustGetUserId(ctx)

	if *userId == followUserId {
		c.Send(ctx).MixedError(NewSelfActionError("follow"))
		return
	}

	err := c.userService.FollowUser(*userId, followUserId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Followed user successfully")
}

func (c *userController) UnfollowUser(ctx *gin.Context) {
	unfollowUserId := ctx.Param("userId")
	userId := c.ContextPayload.MustGetUserId(ctx)

	if *userId == unfollowUserId {
		c.Send(ctx).MixedError(NewSelfActionError("unfollow"))
		return
	}

	err := c.userService.UnfollowUser(*userId, unfollowUserId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Unfollowed user successfully")
}

func (c *userController) BlockUser(ctx *gin.Context) {
	blockUserId := ctx.Param("userId")
	userId := c.ContextPayload.MustGetUserId(ctx)

	if *userId == blockUserId {
		c.Send(ctx).MixedError(NewSelfActionError("block"))
		return
	}

	err := c.userService.BlockUser(*userId, blockUserId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Blocked user successfully")
}

func (c *userController) UnblockUser(ctx *gin.Context) {
	unblockUserId := ctx.Param("userId")
	userId := c.ContextPayload.MustGetUserId(ctx)

	if *userId == unblockUserId {
		c.Send(ctx).MixedError(NewSelfActionError("unblock"))
		return
	}

	err := c.userService.UnblockUser(*userId, unblockUserId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Unblocked user successfully")
}
