package comment

import (
	"fmt"
	"sync-backend/api/comment/dto"
	"sync-backend/api/comment/model"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	community "sync-backend/api/community/model"
	post "sync-backend/api/post/model"
)

type CommentService interface {
	CreatePostComment(userId string, comment *dto.CreatePostCommentRequest) (*model.Comment, network.ApiError)
	EditPostComment(userId string, commentId string, comment *dto.EditPostCommentRequest) (*model.Comment, network.ApiError)
	DeletePostComment(userId string, commentId string) network.ApiError
	GetPostComments(userId string, postId string, page int, limit int) ([]*model.Comment, network.ApiError)
	// GetPostCommentReplies(userId string, postId string, commentId string, page int, limit int) ([]*model.Comment, network.ApiError)
	CreatePostCommentReply(userId string, comment *dto.CreateCommentReplyRequest) (*model.Comment, network.ApiError)
	EditPostCommentReply(userId string, commentId string, comment *dto.EditCommentReplyRequest) (*model.Comment, network.ApiError)
	DeletePostCommentReply(userId string, commentId string) network.ApiError
}

type commentService struct {
	network.BaseService
	logger                utils.AppLogger
	commentQueryBuilder   mongo.QueryBuilder[model.Comment]
	postQueryBuilder      mongo.QueryBuilder[post.Post]
	communityQueryBuilder mongo.QueryBuilder[community.Community]
}

func NewCommentService(db mongo.Database) CommentService {
	return &commentService{
		BaseService:           network.NewBaseService(),
		logger:                utils.NewServiceLogger("CommentService"),
		commentQueryBuilder:   mongo.NewQueryBuilder[model.Comment](db, model.CommentCollectionName),
		postQueryBuilder:      mongo.NewQueryBuilder[post.Post](db, post.PostCollectionName),
		communityQueryBuilder: mongo.NewQueryBuilder[community.Community](db, community.CommunityCollectionName),
	}
}

func (s *commentService) CreatePostComment(userId string, comment *dto.CreatePostCommentRequest) (*model.Comment, network.ApiError) {
	// check for post existence
	postFilter := bson.M{"postId": comment.PostId}
	_, err := s.postQueryBuilder.SingleQuery().FindOne(postFilter, nil)
	if err != nil {
		s.logger.Error("Failed to find post - %v", err)
		return nil, network.NewNotFoundError("Post not found", fmt.Errorf("post %s not found", comment.PostId))
	}
	// check for community existence
	communityFilter := bson.M{"communityId": comment.CommunityId}
	_, err = s.communityQueryBuilder.SingleQuery().FindOne(communityFilter, nil)
	if err != nil {
		s.logger.Error("Failed to find community - %v", err)
		return nil, network.NewNotFoundError("Community not found", fmt.Errorf("community %s not found", comment.CommunityId))
	}

	commentModel := model.NewComment(comment.PostId, userId, comment.CommunityId, comment.Comment, comment.ParentId)
	commentModel.AddDeviceInfo(comment.DeviceId, comment.DeviceType, comment.DeviceOS, comment.DeviceVersion)
	commentModel.AddLocationInfo(comment.Country, comment.City, comment.Latitude, comment.Longitude, comment.IpAddress, comment.TimeZone)
	_, err = s.commentQueryBuilder.SingleQuery().InsertOne(commentModel)
	if err != nil {
		s.logger.Error("Failed to create post comment - %v", err)
		return nil, network.NewInternalServerError("Failed to create comment", network.DB_ERROR, err)
	}
	return commentModel, nil
}

func (s *commentService) EditPostComment(userId string, commentId string, comment *dto.EditPostCommentRequest) (*model.Comment, network.ApiError) {
	filter := bson.M{"commentId": commentId}
	if comment.ParentId != "" {
		filter["parentId"] = comment.ParentId
	}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(filter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return nil, network.NewNotFoundError("Comment not found", fmt.Errorf("comment %s not found", commentId))
	}

	if commentModel.AuthorId != userId {
		s.logger.Error("User is not authorized to edit this comment")
		return nil, network.NewForbiddenError("User is not authorized to edit this comment", fmt.Errorf("user %s is not authorized to edit comment %s", userId, commentId))
	}

	commentModel.Content = comment.Comment
	commentModel.ParentId = comment.ParentId
	update := bson.M{
		"$set": bson.M{
			"status":    model.CommentStatusActive,
			"isEdited":  true,
			"content":   commentModel.Content,
			"parentId":  commentModel.ParentId,
			"updatedAt": commentModel.UpdatedAt,
		},
		"$inc": bson.M{
			"metadata.version": 1,
		},
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(filter, update, nil)
	if err != nil {
		s.logger.Error("Failed to update post comment - %v", err)
		return nil, network.NewInternalServerError("Failed to update comment", network.DB_ERROR, err)
	}

	return commentModel, nil
}

func (s *commentService) DeletePostComment(userId string, commentId string) network.ApiError {
	filter := bson.M{"commentId": commentId}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(filter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return network.NewNotFoundError("Comment not found", fmt.Errorf("comment %s not found", commentId))
	}
	if commentModel.AuthorId != userId {
		s.logger.Error("User is not authorized to delete this comment")
		return network.NewForbiddenError("User is not authorized to delete this comment", fmt.Errorf("user %s is not authorized to delete comment %s", userId, commentId))
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"commentId": commentId},
		bson.M{
			"$set": bson.M{
				"status":             model.CommentStatusDeleted,
				"isDeleted":          true,
				"deletedAt":          primitive.NewDateTimeFromTime(time.Now()),
				"updatedAt":          primitive.NewDateTimeFromTime(time.Now()),
				"metadata.deletedBy": userId,
				"metadata.version":   commentModel.Metadata.Version + 1,
			},
		},
		nil,
	)
	if err != nil {
		s.logger.Error("Failed to delete post comment - %v", err)
		return network.NewInternalServerError("Failed to delete comment", network.DB_ERROR, err)
	}
	return nil
}

func (s *commentService) GetPostComments(userId string, postId string, page int, limit int) ([]*model.Comment, network.ApiError) {
	filter := bson.M{
		"postId":    postId,
		"isDeleted": false,
		"status":    model.CommentStatusActive,
		"parentId":  bson.M{"$exists": false},
		"path":      postId,
	}
	opts := options.FindOptions{
		Sort: bson.D{
			{Key: "createdAt", Value: -1},
			{Key: "synergy", Value: 1},
		},
	}
	comments, err := s.commentQueryBuilder.SingleQuery().FilterPaginated(filter, int64(page), int64(limit), &opts)
	if err != nil {
		s.logger.Error("Failed to get post comments - %v", err)
		return nil, network.NewInternalServerError("Failed to get comments", network.DB_ERROR, err)
	}
	if len(comments) == 0 {
		return []*model.Comment{}, nil
	} else {
		return comments, nil
	}
}

func (s *commentService) CreatePostCommentReply(userId string, comment *dto.CreateCommentReplyRequest) (*model.Comment, network.ApiError) {
	commentFilter := bson.M{"commentId": comment.CommentId}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(commentFilter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return nil, network.NewNotFoundError("Comment not found", fmt.Errorf("comment %s not found", comment.CommentId))
	}

	if commentModel.Status != model.CommentStatusActive {
		s.logger.Error("Comment is not active")
		return nil, network.NewForbiddenError("Comment is not active", fmt.Errorf("comment %s is not active", comment.CommentId))
	}

	if commentModel.IsDeleted {
		s.logger.Error("Comment is deleted")
		return nil, network.NewForbiddenError("Comment is deleted", fmt.Errorf("comment %s is deleted", comment.CommentId))
	}

	replyComment := model.NewComment(commentModel.PostId, userId, commentModel.CommunityId, comment.Reply, comment.CommentId)
	replyComment.AddDeviceInfo(comment.DeviceId, comment.DeviceType, comment.DeviceOS, comment.DeviceVersion)
	replyComment.AddLocationInfo(comment.Country, comment.City, comment.Latitude, comment.Longitude, comment.IpAddress, comment.TimeZone)
	replyComment.Path = fmt.Sprintf("%s.%s", commentModel.Path, commentModel.CommentId)
	replyComment.ParentId = commentModel.CommentId

	_, err = s.commentQueryBuilder.SingleQuery().InsertOne(replyComment)
	if err != nil {
		s.logger.Error("Failed to create post comment reply - %v", err)
		return nil, network.NewInternalServerError("Failed to create comment reply", network.DB_ERROR, err)
	}

	// update the comment with the new reply
	update := bson.M{
		"$set": bson.M{
			"status":           model.CommentStatusActive,
			"updatedAt":        replyComment.UpdatedAt,
			"metadata.version": replyComment.Metadata.Version + 1,
		},
		"$inc": bson.M{
			"replyCount": 1,
		},
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(commentFilter, update, nil)
	if err != nil {
		s.logger.Error("Failed to update post comment with reply - %v", err)
		return nil, network.NewInternalServerError("Failed to update comment with reply", network.DB_ERROR, err)
	}

	return replyComment, nil
}

func (s *commentService) EditPostCommentReply(userId string, commentId string, comment *dto.EditCommentReplyRequest) (*model.Comment, network.ApiError) {
	filter := bson.M{"commentId": commentId}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(filter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return nil, network.NewNotFoundError("Comment not found", fmt.Errorf("comment %s not found", commentId))
	}

	if commentModel.AuthorId != userId {
		s.logger.Error("User is not authorized to edit this comment")
		return nil, network.NewForbiddenError("User is not authorized to edit this comment", fmt.Errorf("user %s is not authorized to edit comment %s", userId, commentId))
	}

	commentModel.Content = comment.Reply
	commentModel.ParentId = comment.CommentId
	update := bson.M{
		"$set": bson.M{
			"status":    model.CommentStatusActive,
			"isEdited":  true,
			"content":   commentModel.Content,
			"parentId":  commentModel.ParentId,
			"updatedAt": commentModel.UpdatedAt,
		},
		"$inc": bson.M{
			"metadata.version": 1,
		},
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(filter, update, nil)
	if err != nil {
		s.logger.Error("Failed to update post comment reply - %v", err)
		return nil, network.NewInternalServerError("Failed to update comment reply", network.DB_ERROR, err)
	}

	return commentModel, nil
}

func (s *commentService) DeletePostCommentReply(userId string, commentId string) network.ApiError {
	filter := bson.M{"commentId": commentId}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(filter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return network.NewNotFoundError("Comment not found", fmt.Errorf("comment %s not found", commentId))
	}
	if commentModel.AuthorId != userId {
		s.logger.Error("User is not authorized to delete this comment")
		return network.NewForbiddenError("User is not authorized to delete this comment", fmt.Errorf("user %s is not authorized to delete comment %s", userId, commentId))
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"commentId": commentId},
		bson.M{
			"$set": bson.M{
				"status":             model.CommentStatusDeleted,
				"isDeleted":          true,
				"deletedAt":          primitive.NewDateTimeFromTime(time.Now()),
				"updatedAt":          primitive.NewDateTimeFromTime(time.Now()),
				"metadata.deletedBy": userId,
				"metadata.version":   commentModel.Metadata.Version + 1,
			},
		},
		nil,
	)
	if err != nil {
		s.logger.Error("Failed to delete post comment reply - %v", err)
		return network.NewInternalServerError("Failed to delete comment reply", network.DB_ERROR, err)
	}
	
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"commentId": commentModel.ParentId},
		bson.M{
			"$set": bson.M{
				"status":             model.CommentStatusActive,
				"updatedAt":          primitive.NewDateTimeFromTime(time.Now()),
				"metadata.version":   commentModel.Metadata.Version + 1,
			},
			"$inc": bson.M{
				"replyCount": -1,
			},
		},
		nil,
	)
	if err != nil {
		s.logger.Error("Failed to update post comment reply - %v", err)
		return network.NewInternalServerError("Failed to update comment reply", network.DB_ERROR, err)
	}
	return nil
}