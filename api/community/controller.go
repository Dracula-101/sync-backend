package community

import (
	"sync-backend/api/community/dto"
	"sync-backend/api/community/model"
	"sync-backend/arch/common"
	coreMW "sync-backend/arch/middleware"
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
	uploadProvider   coreMW.UploadProvider
	communityService CommunityService
}

func NewCommunityController(
	authProvider network.AuthenticationProvider,
	uploadProvider coreMW.UploadProvider,
	communityService CommunityService,
) network.Controller {
	return &communityController{
		logger:           utils.NewServiceLogger("CommunityController"),
		BaseController:   network.NewBaseController("/api/v1/community", authProvider),
		ContextPayload:   common.NewContextPayload(),
		authProvider:     authProvider,
		uploadProvider:   uploadProvider,
		communityService: communityService,
	}
}

func (c *communityController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting community routes")
	group.Use(c.authProvider.Middleware())
	group.POST("/create", c.CreateCommunity)
	group.GET("/:communityId", c.GetCommunityById)
	group.GET("/search", c.SearchCommunities)
	group.GET("/autocomplete", c.AutocompeleteCommunities)
	group.GET("/trending", c.GetTrendingCommunities)
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
	query, err := network.ReqQuery(ctx, dto.NewSearchCommunityRequest())
	if err != nil {
		return
	}

	communities, err := c.communityService.SearchCommunities(
		query.Query,
		query.Page,
		query.Limit,
		query.ShowPrivate,
	)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Communities fetched successfully", communities)
}

func (c *communityController) AutocompeleteCommunities(ctx *gin.Context) {
	query, err := network.ReqQuery(ctx, dto.NewAutocompleteCommunityRequest())
	if err != nil {
		return
	}

	communities, err := c.communityService.AutocompleteCommunities(query.Query, query.Page, query.Limit, query.ShowPrivate)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Communities fetched successfully", communities)
}

func (c *communityController) GetTrendingCommunities(ctx *gin.Context) {
	query, err := network.ReqQuery(ctx, dto.NewGetTrendingCommunitiesRequest())
	if err != nil {
		return
	}

	communities, err := c.communityService.GetTrendingCommunities(query.Page, query.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Communities fetched successfully", communities)
}
