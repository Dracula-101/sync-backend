package post

import (
	"sync-backend/api/post/dto"
	"sync-backend/api/post/model"
	"sync-backend/arch/common"
	"sync-backend/arch/middleware"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type postController struct {
	network.BaseController
	common.ContextPayload
	authenticatorProvider network.AuthenticationProvider
	uploadProvider        middleware.UploadProvider
	logger                utils.AppLogger
	postService           PostService
}

func NewPostController(postService PostService, authenticatorProvider network.AuthenticationProvider, uploadProvider middleware.UploadProvider) network.Controller {
	return &postController{
		BaseController:        network.NewBaseController("/api/v1/post", authenticatorProvider),
		ContextPayload:        common.NewContextPayload(),
		logger:                utils.NewServiceLogger("PostController"),
		authenticatorProvider: authenticatorProvider,
		uploadProvider:        uploadProvider,
		postService:           postService,
	}
}

func (c *postController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting post routes")
	group.POST("/create", c.authenticatorProvider.Middleware(), c.uploadProvider.Middleware("media"), c.CreatePost)
	group.GET("/get/:postId", c.GetPost)
	group.POST("/edit/:postId", c.authenticatorProvider.Middleware(), c.EditPost)
	group.GET("/get/user", c.authenticatorProvider.Middleware(), c.UserPosts)
	group.GET("/get/community/:communityId", c.GetCommunityPosts)
}

func (c *postController) CreatePost(ctx *gin.Context) {
	body, err := network.ReqForm(ctx, dto.NewCreatePostRequest())
	if err != nil {
		return
	}
	userId := c.MustGetUserId(ctx)
	files := c.uploadProvider.GetUploadedFiles(ctx)

	c.logger.Info("Creating post with title: %s", body.Title)
	var filesList []string
	if files != nil && len(files.Files) > 0 {
		for _, file := range files.Files {
			c.logger.Debug("File uploaded: %s", file.Path)
			filesList = append(filesList, file.Path)
		}
	}
	// TODO: add media service which uploads the images
	post, err := c.postService.CreatePost(
		body.Title,
		body.Content,
		body.Tags,
		filesList,
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
	c.logger.Debug("Post details: %+v", post)
	c.uploadProvider.DeleteUploadedFiles(ctx)
}

func (c *postController) GetPost(ctx *gin.Context) {
	postId := ctx.Param("postId")
	if postId == "" {
		c.Send(ctx).BadRequestError("Post ID is required", nil)
		return
	}
	post, err := c.postService.GetPost(postId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("Post retrieved successfully", post)
	c.logger.Debug("Post details: %+v", post)
}

func (c *postController) EditPost(ctx *gin.Context) {
	postId := ctx.Param("postId")
	if postId == "" {
		c.Send(ctx).BadRequestError("Post ID is required", nil)
		return
	}
	body, err := network.ReqForm(ctx, dto.NewEditPostRequest())
	if err != nil {
		return
	}
	userId := c.MustGetUserId(ctx)
	title := body.Title
	content := body.Content
	isNSFW := body.IsNSFW
	isSpoiler := body.IsSpoiler
	_, err = c.postService.EditPost(
		*userId,
		postId,
		&title,
		&content,
		body.PostType,
		&isNSFW,
		&isSpoiler,
	)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("Post edited successfully")
}

func (c *postController) UserPosts(ctx *gin.Context) {
	userId := c.MustGetUserId(ctx)
	body, err := network.ReqQuery(ctx, dto.NewGetUserPostRequest())
	if err != nil {
		return
	}
	posts, numberPosts, err := c.postService.GetPostsByUserId(*userId, body.Page, body.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	postsValue := make([]model.Post, len(posts))
	for i, post := range posts {
		if post != nil {
			postsValue[i] = *post
		}
	}
	c.Send(ctx).SuccessDataResponse("User posts retrieved successfully", dto.NewGetUserPostResponse(postsValue, body.Page, body.Limit, numberPosts))
}

func (c *postController) GetCommunityPosts(ctx *gin.Context) {
	communityId := ctx.Param("communityId")
	if communityId == "" {
		c.Send(ctx).BadRequestError("Community ID is required", nil)
		return
	}
	body, err := network.ReqQuery(ctx, dto.NewGetCommunityPostRequest())
	if err != nil {
		return
	}
	posts, numberPosts, err := c.postService.GetPostsByCommunityId(communityId, body.Page, body.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	postsValue := make([]model.Post, len(posts))
	for i, post := range posts {
		if post != nil {
			postsValue[i] = *post
		}
	}
	c.Send(ctx).SuccessDataResponse("Community posts retrieved successfully", dto.NewGetCommunityPostResponse(postsValue, body.Page, body.Limit, numberPosts))
}
