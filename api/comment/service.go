package comment

import (
	"fmt"
	"sync-backend/api/comment/dto"
	"sync-backend/api/comment/model"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"go.mongodb.org/mongo-driver/bson"

	community "sync-backend/api/community/model"
	post "sync-backend/api/post/model"
)

type CommentService interface {
	CreatePostComment(userId string, comment *dto.CreatePostCommentRequest) (*model.Comment, network.ApiError)
	EditPostComment(userId string, commentId string, comment *dto.EditPostCommentRequest) (*model.Comment, network.ApiError)
	DeletePostComment(userId string, commentId string) network.ApiError
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
	commentModel.AddLocationInfo(
		comment.Country,
		comment.City,
		comment.Latitude,
		comment.Longitude,
		comment.IpAddress,
		comment.TimeZone,
	)
	_, err = s.commentQueryBuilder.SingleQuery().InsertOne(commentModel)
	if err != nil {
		s.logger.Error("Failed to create post comment - %v", err)
		return nil, network.NewInternalServerError("Failed to create comment", network.DB_ERROR, err)
	}
	return commentModel, nil
}

func (s *commentService) EditPostComment(userId string, commentId string, comment *dto.EditPostCommentRequest) (*model.Comment, network.ApiError) {
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

	commentModel.Content = comment.Comment
	commentModel.ParentId = comment.ParentId
	update := bson.M{
		"$set": bson.M{
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
				"deleted":          true,
				"deletedAt":        commentModel.DeletedAt,
				"deletedBy":        userId,
				"updatedAt":        commentModel.UpdatedAt,
				"metadata.version": commentModel.Metadata.Version + 1,
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
