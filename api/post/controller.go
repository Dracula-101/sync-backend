package post

import (
	"sync-backend/api/post/dto"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type postController struct {
	network.BaseController
	common.ContextPayload
	authenticatorProvider network.AuthenticationProvider
	logger                utils.AppLogger
	postService           PostService
}

func NewPostController(postService PostService, authenticatorProvider network.AuthenticationProvider) network.Controller {
	return &postController{
		BaseController:        network.NewBaseController("/api/v1/post", authenticatorProvider),
		ContextPayload:        common.NewContextPayload(),
		logger:                utils.NewServiceLogger("PostController"),
		authenticatorProvider: authenticatorProvider,
		postService:           postService,
	}
}

func (c *postController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting post routes")
	group.POST("/create", c.authenticatorProvider.Middleware(), c.CreatePost)
}

func (c *postController) CreatePost(ctx *gin.Context) {
	body, err := network.ReqForm(ctx, dto.NewCreatePostRequest())
	if err != nil {
		return
	}
	userId := c.MustGetUserId(ctx)

	c.logger.Info("Creating post with title: %s", body.Title)
	post, err := c.postService.CreatePost(
		body.Title,
		body.Content,
		body.Tags,
		body.Media,
		*userId,
		body.CommunityId,
		body.Type,
		body.IsNSFW,
		body.IsSpoiler,
	)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Post created successfully", dto.CreatePostResponse{PostId: post.PostId})
	c.logger.Info("Post created successfully with ID: %s", post.PostId)
	c.logger.Debug("Post details: %+v", post)
}
