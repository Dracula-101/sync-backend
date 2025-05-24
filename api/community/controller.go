package community

import (
	"strings"
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
		BaseController:   network.NewBaseController("/community", authProvider),
		ContextPayload:   common.NewContextPayload(),
		authProvider:     authProvider,
		uploadProvider:   uploadProvider,
		communityService: communityService,
	}
}

func (c *communityController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting community routes")
	group.Use(c.authProvider.Middleware())
	group.POST("/create", c.uploadProvider.Middleware("avatar_photo", "background_photo"), c.CreateCommunity)
	group.GET("/:communityId", c.GetCommunityById)
	group.PUT("/:communityId", c.uploadProvider.Middleware("avatar_photo", "background_photo"), c.UpdateCommunity)
	group.DELETE("/:communityId", c.DeleteCommunity)

	group.GET("/search", c.SearchCommunities)
	group.GET("/autocomplete", c.AutocompeleteCommunities)
	group.GET("/trending", c.GetTrendingCommunities)
}

func (c *communityController) CreateCommunity(ctx *gin.Context) {
	body, err := network.ReqForm(ctx, dto.NewCreateCommunityRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	avatarPhoto := c.uploadProvider.GetUploadedFiles(ctx, "avatar_photo")
	backgroundPhoto := c.uploadProvider.GetUploadedFiles(ctx, "background_photo")
	if len(avatarPhoto.Files) > 0 {
		body.AvatarFilePath = avatarPhoto.Files[0].Path
	}
	if len(backgroundPhoto.Files) > 0 {
		body.BackgroundFilePath = backgroundPhoto.Files[0].Path
	}

	// Process the tagIds - already validated through custom validator
	tagIds := strings.Split(body.TagIds, ",")
	for i := range tagIds {
		tagIds[i] = strings.TrimSpace(tagIds[i])
	}

	community, err := c.communityService.CreateCommunity(body.Name, body.Description, tagIds, body.AvatarFilePath, body.BackgroundFilePath, *userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Community created successfully", dto.CreateCommunityResponse{
		CommunityId: community.CommunityId,
		Name:        community.Name,
		Slug:        community.Slug,
	})
	c.uploadProvider.DeleteUploadedFiles(ctx, "avatar_photo")
	c.uploadProvider.DeleteUploadedFiles(ctx, "background_photo")
}

func (c *communityController) UpdateCommunity(ctx *gin.Context) {
	body, err := network.ReqForm(ctx, dto.NewUpdateCommunityRequest())
	if err != nil {
		return
	}

	userId := c.ContextPayload.MustGetUserId(ctx)
	avatarPhoto := c.uploadProvider.GetUploadedFiles(ctx, "avatar_photo")
	backgroundPhoto := c.uploadProvider.GetUploadedFiles(ctx, "background_photo")
	if len(avatarPhoto.Files) > 0 {
		body.AvatarFilePath = avatarPhoto.Files[0].Path
	}
	if len(backgroundPhoto.Files) > 0 {
		body.BackgroundFilePath = backgroundPhoto.Files[0].Path
	}

	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			err,
		)
		return
	}

	_, err = c.communityService.UpdateCommunity(
		communityId,
		body.CommunityDescription,
		body.AvatarFilePath,
		body.BackgroundFilePath,
		*userId,
	)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("Community updated successfully")
	c.uploadProvider.DeleteUploadedFiles(ctx, "avatar_photo")
	c.uploadProvider.DeleteUploadedFiles(ctx, "background_photo")
}

func (c *communityController) DeleteCommunity(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError(
			"Community ID is required",
			"Please provide a valid community ID in the request params",
			nil,
		)
		return
	}
	userId := c.ContextPayload.MustGetUserId(ctx)
	err := c.communityService.DeleteCommunity(communityId, *userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("Community deleted successfully")
	c.uploadProvider.DeleteUploadedFiles(ctx, "avatar_photo")
	c.uploadProvider.DeleteUploadedFiles(ctx, "background_photo")
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
