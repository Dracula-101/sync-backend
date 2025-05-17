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
	locationProvider      network.LocationProvider
	logger                utils.AppLogger
	commentService        CommentService
}

func NewCommentController(authenticatorProvider network.AuthenticationProvider, locationProvider network.LocationProvider, commentService CommentService) *commentController {
	return &commentController{
		BaseController:        network.NewBaseController("/api/v1/comment", authenticatorProvider),
		ContextPayload:        common.NewContextPayload(),
		logger:                utils.NewServiceLogger("CommentController"),
		authenticatorProvider: authenticatorProvider,
		locationProvider:      locationProvider,
		commentService:        commentService,
	}
}

func (c *commentController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting comment routes")
	group.Use(c.authenticatorProvider.Middleware())
	group.POST("/post/create", c.locationProvider.Middleware(), c.CreatePostComment)
	group.POST("/post/edit/:commentId", c.EditPostComment)
	group.POST("/post/delete/:commentId", c.DeletePostComment)
	group.GET("/post/:postId", c.GetPostComments)
	group.GET("/post/:postId/reply/:commentId", c.GetPostCommentReplies)

	group.POST("/post/reply/create", c.locationProvider.Middleware(), c.CreatePostCommentReply)
	group.POST("/post/reply/edit/:commentId", c.EditPostCommentReply)
	group.POST("/post/reply/delete/:commentId", c.DeletePostCommentReply)

	group.POST("/like/:commentId", c.LikePostComment)
	group.POST("/dislike/:commentId", c.DislikePostComment)
}

func (c *commentController) CreatePostComment(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewCreatePostCommentRequest())
	if err != nil {
		return
	}

	userId := c.MustGetUserId(ctx)
	c.SetRequestDeviceDetails(ctx, &body.BaseDeviceRequest)
	c.SetRequestLocationDetails(ctx, &body.BaseLocationRequest)
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
	if commentId == "" {
		c.logger.Error("Comment ID is required")
		c.Send(ctx).BadRequestError("Comment ID is required", nil)
		return
	}

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

func (c *commentController) GetPostComments(ctx *gin.Context) {
	postId := ctx.Param("postId")
	if postId == "" {
		c.logger.Error("Post ID is required")
		c.Send(ctx).BadRequestError("Post ID is required", nil)
		return
	}
	params, err := network.ReqQuery(ctx, dto.NewGetPostComentRequest())
	if err != nil {
		c.logger.Error("Failed to parse query parameters: %v", err)
		return
	}
	userId := c.MustGetUserId(ctx)
	comments, err := c.commentService.GetPostComments(*userId, postId, params.Pagination.Page, params.Pagination.Limit)
	if err != nil {
		c.logger.Error("Failed to get post comments: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Comments retrieved successfully", comments)
}

func (c *commentController) GetPostCommentReplies(ctx *gin.Context) {
	commentId := ctx.Param("commentId")
	if commentId == "" {
		c.logger.Error("Comment ID is required")
		c.Send(ctx).BadRequestError("Comment ID is required", nil)
		return
	}
	postId := ctx.Param("postId")
	if postId == "" {
		c.logger.Error("Post ID is required")
		c.Send(ctx).BadRequestError("Post ID is required", nil)
		return
	}

	params, err := network.ReqQuery(ctx, dto.NewGetPostRepliesParams())
	if err != nil {
		c.logger.Error("Failed to parse query parameters: %v", err)
		return
	}

	userId := c.MustGetUserId(ctx)
	replies, err := c.commentService.GetPostCommentReplies(*userId, postId, commentId, params.Page, params.Limit)
	if err != nil {
		c.logger.Error("Failed to get post comment replies: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Replies retrieved successfully", replies)
}

func (c *commentController) CreatePostCommentReply(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewCreateCommentReplyRequest())
	if err != nil {
		return
	}

	userId := c.MustGetUserId(ctx)
	c.SetRequestDeviceDetails(ctx, &body.BaseDeviceRequest)
	c.SetRequestLocationDetails(ctx, &body.BaseLocationRequest)

	_, err = c.commentService.CreatePostCommentReply(*userId, body)
	if err != nil {
		c.logger.Error("Failed to create post comment reply: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Comment reply created successfully")
}

func (c *commentController) EditPostCommentReply(ctx *gin.Context) {
	commentId := ctx.Param("commentId")
	if commentId == "" {
		c.logger.Error("Comment ID is required")
		c.Send(ctx).BadRequestError("Comment ID is required", nil)
		return
	}

	body, err := network.ReqBody(ctx, dto.NewEditCommentReplyRequest())
	if err != nil {
		return
	}

	userId := c.MustGetUserId(ctx)
	_, err = c.commentService.EditPostCommentReply(*userId, commentId, body)
	if err != nil {
		c.logger.Error("Failed to edit post comment reply: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("Comment reply edited successfully")

}

func (c *commentController) DeletePostCommentReply(ctx *gin.Context) {
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

func (c *commentController) LikePostComment(ctx *gin.Context) {
	commentId := ctx.Param("commentId")
	userId := c.MustGetUserId(ctx)

	isLiked, synergy, err := c.commentService.LikePostComment(*userId, commentId)
	if err != nil {
		c.logger.Error("Failed to like post comment: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Comment liked successfully", dto.NewLikePostCommentResponse(*isLiked, *synergy))
}

func (c *commentController) DislikePostComment(ctx *gin.Context) {
	commentId := ctx.Param("commentId")
	userId := c.MustGetUserId(ctx)

	isDisliked, synergy, err := c.commentService.DislikePostComment(*userId, commentId)
	if err != nil {
		c.logger.Error("Failed to dislike post comment: %v", err)
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Comment disliked successfully", dto.NewDislikePostCommentResponse(*isDisliked, *synergy))
}
