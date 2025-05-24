package user

import (
	"fmt"
	"sync-backend/api/common/location"
	"sync-backend/api/community"
	communityModel "sync-backend/api/community/model"
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
	authProvider     network.AuthenticationProvider
	uploadProvider   coreMW.UploadProvider
	userService      UserService
	communityService community.CommunityService
	locationService  location.LocationService
}

func NewUserController(
	authProvider network.AuthenticationProvider,
	uploadProvider coreMW.UploadProvider,
	userService UserService,
	communityService community.CommunityService,
	locationService location.LocationService,
) network.Controller {
	return &userController{
		logger:           utils.NewServiceLogger("UserController"),
		BaseController:   network.NewBaseController("/user", nil),
		ContextPayload:   common.NewContextPayload(),
		authProvider:     authProvider,
		uploadProvider:   uploadProvider,
		userService:      userService,
		communityService: communityService,
		locationService:  locationService,
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

	// User community routes
	group.POST("/join/:communityId", c.JoinCommunity)
	group.POST("/leave/:communityId", c.LeaveCommunity)
	group.GET("/communities/owner", c.GetMyCommunities)
	group.GET("/communities/joined", c.GetJoinedCommunities)
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
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", *userId),
			nil,
		)
		return
	}

	err = c.communityService.JoinCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
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
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", *userId),
			nil,
		)
		return
	}

	err = c.communityService.LeaveCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	err = c.userService.LeaveCommunity(*userId, communityId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Left community successfully")
}

func (c *userController) GetMyCommunities(ctx *gin.Context) {
	body, err := network.ReqQuery(ctx, dto.NewGetMyCommunitiesRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	communities, err := c.communityService.GetCommunities(*userId, body.Page, body.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.logger.Debug("Communities: %+v", communities)

	if communities == nil {
		c.Send(ctx).NotFoundError(
			"No communities found",
			"No communities found for the user",
			nil,
		)
		return
	}

	var finalCommunities []communityModel.Community
	for _, community := range communities {
		if community != nil {
			finalCommunities = append(finalCommunities, *community)
		}
	}

	c.Send(ctx).SuccessDataResponse(
		"Communities fetched successfully",
		dto.NewGetMyCommunitiesResponse(
			finalCommunities,
			len(communities),
		),
	)

}

func (c *userController) GetJoinedCommunities(ctx *gin.Context) {
	body, err := network.ReqQuery(ctx, dto.NewJoinedCommunitiesRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	communities, err := c.communityService.GetCommunities(*userId, body.Page, body.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if communities == nil {
		c.Send(ctx).NotFoundError(
			"No communities found",
			"No communities found for the user",
			nil,
		)
		return
	}

	var finalCommunities []communityModel.Community
	for _, community := range communities {
		if community != nil {
			finalCommunities = append(finalCommunities, *community)
		}
	}

	c.Send(ctx).SuccessDataResponse(
		"Communities fetched successfully",
		dto.NewJoinedCommunitiesResponse(
			finalCommunities,
			len(communities),
		),
	)
}
