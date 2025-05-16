package comment

import (
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"

	"sync-backend/api/comment/dto"
)

type commentController struct {
	network.BaseController
	common.ContextPayload
	authenticatorProvider network.AuthenticationProvider
	logger                utils.AppLogger
	commentService        CommentService
}

func NewCommentController(authenticatorProvider network.AuthenticationProvider, commentService CommentService) network.Controller {
	return &commentController{
		BaseController:        network.NewBaseController("/api/v1/comment", authenticatorProvider),
		ContextPayload:        common.NewContextPayload(),
		logger:                utils.NewServiceLogger("CommentController"),
		authenticatorProvider: authenticatorProvider,
		commentService:        commentService,
	}
}

func (c *commentController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting comment routes")
	group.Use(c.authenticatorProvider.Middleware())
	group.POST("/post/create", c.CreatePostComment)
	group.POST("/post/edit/:commentId", c.EditPostComment)
	group.POST("/post/delete/:commentId", c.DeletePostComment)
}

func (c *commentController) CreatePostComment(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewCreatePostCommentRequest())
	if err != nil {
		return
	}

	userId := c.MustGetUserId(ctx)
	_, err = c.commentService.CreatePostComment(*userId, body)
	if err != nil {
		c.logger.Error("Failed to create post comment: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Comment created successfully")
}

func (c *commentController) EditPostComment(ctx *gin.Context) {
	commentId := ctx.Param("commentId")
	body, err := network.ReqBody(ctx, dto.NewEditPostCommentRequest())
	if err != nil {
		return
	}

	userId := c.MustGetUserId(ctx)
	_, err = c.commentService.EditPostComment(*userId, commentId, body)
	if err != nil {
		c.logger.Error("Failed to edit post comment: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Comment edited successfully")
}

func (c *commentController) DeletePostComment(ctx *gin.Context) {
	commentId := ctx.Param("commentId")
	userId := c.MustGetUserId(ctx)

	err := c.commentService.DeletePostComment(*userId, commentId)
	if err != nil {
		c.logger.Error("Failed to delete post comment: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Comment deleted successfully")
}
