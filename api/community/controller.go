package community

import (
	"sync-backend/api/community/dto"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type CommunityController interface {
	CreateCommunity(request *dto.CreateCommunityRequest) (*dto.CreateCommunityResponse, error)
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
}

func (c *communityController) CreateCommunity(ctx *gin.Context) {
	_, err := network.ReqBody(ctx, dto.NewCreateCommunityRequest())
	if err != nil {
		c.Send(ctx).BadRequestError("Invalid request body", err)
		return
	}
}
