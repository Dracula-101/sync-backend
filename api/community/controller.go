package community

import (
	"sync-backend/api/community/dto"
	"sync-backend/api/community/model"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type CommunityController interface {
	CreateCommunity(request *dto.CreateCommunityRequest) (*dto.CreateCommunityResponse, error)
	GetCommunityById(id string) (*model.Community, error)
	SearchCommunities(query string) ([]model.Community, error)
	GetMyCommunities(userId string) ([]model.Community, error)
}

type communityController struct {
	logger utils.AppLogger
	network.BaseController
	common.ContextPayload
	authProvider     network.AuthenticationProvider
	communityService CommunityService
}

func NewCommunityController(
	communityService CommunityService,
	authProvider network.AuthenticationProvider,
) network.Controller {
	return &communityController{
		logger:           utils.NewServiceLogger("CommunityController"),
		BaseController:   network.NewBaseController("/api/v1/community", authProvider),
		ContextPayload:   common.NewContextPayload(),
		authProvider:     authProvider,
		communityService: communityService,
	}
}

func (c *communityController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting community routes")
	group.POST("/create", c.authProvider.Middleware(), c.CreateCommunity)
	group.GET("/:communityId", c.GetCommunityById)
	group.GET("/search", c.authProvider.Middleware(), c.SearchCommunities)
	group.GET("/my-communities", c.authProvider.Middleware(), c.GetMyCommunities)
}

func (c *communityController) CreateCommunity(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewCreateCommunityRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	community, err := c.communityService.CreateCommunity(body.Name, body.Description, body.TagIds, body.AvatarUrl, body.BackgroundUrl, *userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("Community created successfully", dto.CreateCommunityResponse{
		CommunityId: community.CommunityId,
		Name:        community.Name,
		Slug:        community.Slug,
	})
}

func (c *communityController) GetCommunityById(ctx *gin.Context) {
	params, err := network.ReqParams(ctx, dto.NewGetCommunityRequest())
	if err != nil {
		return
	}

	community, err := c.communityService.GetCommunityById(params.Id)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Community fetched successfully", community)
}

func (c *communityController) SearchCommunities(ctx *gin.Context) {
}

func (c *communityController) GetMyCommunities(ctx *gin.Context) {
}
