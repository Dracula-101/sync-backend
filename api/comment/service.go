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
	GetPostComments(userId string, postId string, page int, limit int) ([]*model.PublicGetComment, network.ApiError)
	GetPostCommentReplies(userId string, postId string, parentId string, page int, limit int) ([]*model.PublicGetComment, network.ApiError)

	CreatePostCommentReply(userId string, comment *dto.CreateCommentReplyRequest) (*model.Comment, network.ApiError)
	EditPostCommentReply(userId string, commentId string, comment *dto.EditCommentReplyRequest) (*model.Comment, network.ApiError)
	DeletePostCommentReply(userId string, commentId string) network.ApiError

	LikePostComment(userId string, commentId string) (*bool, *int, network.ApiError)
	DislikePostComment(userId string, commentId string) (*bool, *int, network.ApiError)

	GetUserComments(userId string, page int, limit int) ([]*model.PublicGetComment, network.ApiError)
}

type commentService struct {
	network.BaseService
	logger                         utils.AppLogger
	commentQueryBuilder            mongo.QueryBuilder[model.Comment]
	commentInteractionQueryBuilder mongo.QueryBuilder[model.CommentInteraction]
	postQueryBuilder               mongo.QueryBuilder[post.Post]
	communityQueryBuilder          mongo.QueryBuilder[community.Community]
	commentAggregateBuilder        mongo.AggregateBuilder[model.Comment, model.PublicGetComment]
	transaction                    mongo.TransactionBuilder
}

func NewCommentService(db mongo.Database) CommentService {
	return &commentService{
		BaseService:                    network.NewBaseService(),
		logger:                         utils.NewServiceLogger("CommentService"),
		commentQueryBuilder:            mongo.NewQueryBuilder[model.Comment](db, model.CommentCollectionName),
		commentInteractionQueryBuilder: mongo.NewQueryBuilder[model.CommentInteraction](db, model.CommentInteractionCollectionName),
		postQueryBuilder:               mongo.NewQueryBuilder[post.Post](db, post.PostCollectionName),
		communityQueryBuilder:          mongo.NewQueryBuilder[community.Community](db, community.CommunityCollectionName),
		commentAggregateBuilder:        mongo.NewAggregateBuilder[model.Comment, model.PublicGetComment](db, model.CommentCollectionName),
		transaction:                    mongo.NewTransactionBuilder(db),
	}
}

func (s *commentService) CreatePostComment(userId string, comment *dto.CreatePostCommentRequest) (*model.Comment, network.ApiError) {
	// check for post existence
	postFilter := bson.M{"postId": comment.PostId}
	_, err := s.postQueryBuilder.SingleQuery().FindOne(postFilter, nil)
	if err != nil {
		s.logger.Error("Failed to find post - %v", err)
		return nil, NewPostNotFoundError(comment.PostId)
	}
	// check for community existence
	communityFilter := bson.M{"communityId": comment.CommunityId}
	_, err = s.communityQueryBuilder.SingleQuery().FindOne(communityFilter, nil)
	if err != nil {
		s.logger.Error("Failed to find community - %v", err)
		return nil, NewCommunityNotFoundError(comment.CommunityId)
	}

	commentModel := model.NewComment(comment.PostId, userId, comment.CommunityId, comment.Comment, comment.ParentId)
	commentModel.AddDeviceInfo(comment.DeviceId, comment.DeviceType, comment.DeviceOS, comment.DeviceVersion)
	commentModel.AddLocationInfo(comment.Country, comment.City, comment.Latitude, comment.Longitude, comment.IpAddress, comment.TimeZone)
	_, err = s.commentQueryBuilder.SingleQuery().InsertOne(commentModel)
	if err != nil {
		s.logger.Error("Failed to create post comment - %v", err)
		return nil, NewDBError("creating comment", err.Error())
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
		return nil, NewCommentNotFoundError(commentId)
	}

	if commentModel.AuthorId != userId {
		s.logger.Error("User is not authorized to edit this comment")
		return nil, NewForbiddenError("edit", userId, commentId)
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
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(filter, update, nil)
	if err != nil {
		s.logger.Error("Failed to update post comment - %v", err)
		return nil, NewDBError("updating comment", err.Error())
	}

	return commentModel, nil
}

func (s *commentService) DeletePostComment(userId string, commentId string) network.ApiError {
	filter := bson.M{"commentId": commentId}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(filter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return NewCommentNotFoundError(commentId)
	}
	if commentModel.AuthorId != userId {
		s.logger.Error("User is not authorized to delete this comment")
		return NewForbiddenError("delete", userId, commentId)
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"commentId": commentId},
		bson.M{
			"$set": bson.M{
				"status":    model.CommentStatusDeleted,
				"isDeleted": true,
				"deletedAt": primitive.NewDateTimeFromTime(time.Now()),
				"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
			},
		},
		nil,
	)
	if err != nil {
		s.logger.Error("Failed to delete post comment - %v", err)
		return NewDBError("deleting comment", err.Error())
	}
	return nil
}

func (s *commentService) GetPostComments(userId string, postId string, page int, limit int) ([]*model.PublicGetComment, network.ApiError) {
	s.logger.Debug("GetPostComments - postId: %s, page: %d, limit: %d", postId, page, limit)
	aggregate := s.commentAggregateBuilder.SingleAggregate()
	aggregate.Match(bson.M{"postId": postId, "status": model.CommentStatusActive, "isDeleted": false, "parentId": bson.M{"$exists": false}})
	aggregate.Sort(bson.D{{Key: "createdAt", Value: -1}, {Key: "synergy", Value: -1}})
	aggregate.Limit(int64(limit))
	aggregate.Skip(int64((page - 1) * limit))
	aggregate.Lookup("users", "authorId", "userId", "author")
	aggregate.Lookup("communities", "communityId", "communityId", "community")

	// Lookup user's interaction with these comments if userId is provided
	if userId != "" {
		aggregate.Lookup(
			model.CommentInteractionCollectionName,
			"commentId",
			"commentId",
			"interactions",
		)
	}

	aggregate.AddFields(bson.M{
		"author":    bson.M{"$arrayElemAt": bson.A{"$author", 0}},
		"community": bson.M{"$arrayElemAt": bson.A{"$community", 0}},
	})

	// If userId is provided, filter interactions for this specific user
	if userId != "" {
		aggregate.AddFields(bson.M{
			"userInteractions": bson.M{
				"$filter": bson.M{
					"input": "$interactions",
					"as":    "interaction",
					"cond": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$$interaction.userId", userId}},
							bson.M{"$in": bson.A{
								"$$interaction.interactionType",
								bson.A{model.CommentInteractionTypeLike, model.CommentInteractionTypeDislike},
							}},
						},
					},
				},
			},
		})
		aggregate.AddFields(bson.M{
			"userInteraction": bson.M{"$arrayElemAt": bson.A{"$userInteractions", 0}},
		})
	}

	// Project fields including conditional isLiked/isDisliked fields when userId is provided
	projectFields := bson.M{
		"id":       "$commentId",
		"postId":   1,
		"parentId": 1,
		"author": bson.M{
			"userId":     "$author.userId",
			"username":   "$author.username",
			"email":      "$author.email",
			"avatar":     "$author.avatar.profile.url",
			"background": "$author.avatar.background.url",
			"status":     "$author.status",
		},
		"community": bson.M{
			"id":          "$community.communityId",
			"name":        "$community.name",
			"description": "$community.description",
			"avatar":      "$community.media.avatar.url",
			"background":  "$community.media.background.url",
			"createdAt":   "$community.createdAt",
			"status":      "$community.status",
		},
		"content":          1,
		"formattedContent": 1,
		"status":           1,
		"synergy":          1,
		"replyCount":       1,
		"reactionCounts":   1,
		"level":            1,
		"isEdited":         1,
		"isPinned":         1,
		"isStickied":       1,
		"isLocked":         1,
		"isDeleted":        1,
		"isRemoved":        1,
		"hasMedia":         1,
		"mentions":         1,
		"path":             1,
		"createdAt":        1,
	}

	if userId != "" {
		projectFields["isLiked"] = bson.M{
			"$cond": bson.A{
				bson.M{"$and": bson.A{
					bson.M{"$ifNull": bson.A{"$userInteraction", false}},
					bson.M{"$eq": bson.A{"$userInteraction.interactionType", model.CommentInteractionTypeLike}},
				}},
				true,
				false,
			},
		}
		projectFields["isDisliked"] = bson.M{
			"$cond": bson.A{
				bson.M{"$and": bson.A{
					bson.M{"$ifNull": bson.A{"$userInteraction", false}},
					bson.M{"$eq": bson.A{"$userInteraction.interactionType", model.CommentInteractionTypeDislike}},
				}},
				true,
				false,
			},
		}
	}

	aggregate.Project(projectFields)

	comments, err := aggregate.Exec()
	if err != nil {
		s.logger.Error("Failed to get post comments - %v", err)
		return nil, network.NewInternalServerError(
			"Failed to get comments",
			fmt.Sprintf("It seems the comments for post '%s' could not be retrieved - Aggregation failed. Please try again later. [Context: postId=%s]", postId, postId),
			network.DB_ERROR,
			err,
		)

	}
	if len(comments) == 0 {
		return []*model.PublicGetComment{}, nil
	} else {
		return comments, nil
	}
}

func (s *commentService) GetPostCommentReplies(userId string, postId string, parentId string, page int, limit int) ([]*model.PublicGetComment, network.ApiError) {
	s.logger.Debug("GetPostComments - postId: %s, page: %d, limit: %d", postId, page, limit)
	aggregate := s.commentAggregateBuilder.SingleAggregate()
	aggregate.Match(bson.M{"postId": postId, "status": model.CommentStatusActive, "isDeleted": false, "parentId": parentId})
	aggregate.Sort(bson.D{{Key: "createdAt", Value: -1}, {Key: "synergy", Value: -1}})
	aggregate.Limit(int64(limit))
	aggregate.Skip(int64((page - 1) * limit))
	aggregate.Lookup("users", "authorId", "userId", "author")
	aggregate.Lookup("communities", "communityId", "communityId", "community")

	// Lookup user's interaction with these comment replies if userId is provided
	if userId != "" {
		aggregate.Lookup(
			model.CommentInteractionCollectionName,
			"commentId",
			"commentId",
			"interactions",
		)
	}

	aggregate.AddFields(bson.M{
		"author":    bson.M{"$arrayElemAt": bson.A{"$author", 0}},
		"community": bson.M{"$arrayElemAt": bson.A{"$community", 0}},
	})

	// If userId is provided, filter interactions for this specific user
	if userId != "" {
		aggregate.AddFields(bson.M{
			"userInteractions": bson.M{
				"$filter": bson.M{
					"input": "$interactions",
					"as":    "interaction",
					"cond": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$$interaction.userId", userId}},
							bson.M{"$in": bson.A{
								"$$interaction.interactionType",
								bson.A{model.CommentInteractionTypeLike, model.CommentInteractionTypeDislike},
							}},
						},
					},
				},
			},
		})
		aggregate.AddFields(bson.M{
			"userInteraction": bson.M{"$arrayElemAt": bson.A{"$userInteractions", 0}},
		})
	}

	// Project fields including conditional isLiked/isDisliked fields
	projectFields := bson.M{
		"id":       "$commentId",
		"postId":   1,
		"parentId": 1,
		"author": bson.M{
			"userId":     "$author.userId",
			"username":   "$author.username",
			"email":      "$author.email",
			"avatar":     "$author.avatar.profile.url",
			"background": "$author.avatar.background.url",
			"status":     "$author.status",
		},
		"community": bson.M{
			"id":          "$community.communityId",
			"name":        "$community.name",
			"description": "$community.description",
			"avatar":      "$community.media.avatar.url",
			"background":  "$community.media.background.url",
			"createdAt":   "$community.createdAt",
			"status":      "$community.status",
		},
		"content":          1,
		"formattedContent": 1,
		"status":           1,
		"synergy":          1,
		"replyCount":       1,
		"reactionCounts":   1,
		"level":            1,
		"isEdited":         1,
		"isPinned":         1,
		"isStickied":       1,
		"isLocked":         1,
		"isDeleted":        1,
		"isRemoved":        1,
		"hasMedia":         1,
		"mentions":         1,
		"path":             1,
		"createdAt":        1,
	}

	if userId != "" {
		projectFields["isLiked"] = bson.M{
			"$cond": bson.A{
				bson.M{"$and": bson.A{
					bson.M{"$ifNull": bson.A{"$userInteraction", false}},
					bson.M{"$eq": bson.A{"$userInteraction.interactionType", model.CommentInteractionTypeLike}},
				}},
				true,
				false,
			},
		}
		projectFields["isDisliked"] = bson.M{
			"$cond": bson.A{
				bson.M{"$and": bson.A{
					bson.M{"$ifNull": bson.A{"$userInteraction", false}},
					bson.M{"$eq": bson.A{"$userInteraction.interactionType", model.CommentInteractionTypeDislike}},
				}},
				true,
				false,
			},
		}
	}

	aggregate.Project(projectFields)

	comments, err := aggregate.Exec()
	if err != nil {
		s.logger.Error("Failed to get post comments - %v", err)
		return nil, network.NewInternalServerError(
			"Failed to get comments",
			fmt.Sprintf("It seems the comments for post '%s' could not be retrieved - Aggregation failed. Please try again later. [Context: postId=%s]", postId, postId),
			network.DB_ERROR,
			err,
		)
	}
	if len(comments) == 0 {
		return []*model.PublicGetComment{}, nil
	} else {
		return comments, nil
	}
}

func (s *commentService) CreatePostCommentReply(userId string, comment *dto.CreateCommentReplyRequest) (*model.Comment, network.ApiError) {
	commentFilter := bson.M{"commentId": comment.CommentId}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(commentFilter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return nil, network.NewNotFoundError(
			"Comment not found",
			fmt.Sprintf("Comment with ID '%s' not found, it may have been deleted or the ID is incorrect", comment.CommentId),
			nil,
		)
	}

	switch commentModel.Status {
	case model.CommentStatusDeleted:
		s.logger.Error("Comment is deleted")
		return nil, network.NewForbiddenError(
			"Comment is deleted",
			fmt.Sprintf("Comment with ID '%s' is deleted. It cannot be replied to.", comment.CommentId),
			fmt.Errorf("comment %s is deleted", comment.CommentId),
		)
	case model.CommentStatusHidden:
		s.logger.Error("Comment is hidden")
		return nil, network.NewForbiddenError(
			"Comment is hidden",
			fmt.Sprintf("Comment with ID '%s' is hidden. It cannot be replied to.", comment.CommentId),
			fmt.Errorf("comment %s is hidden", comment.CommentId),
		)
	case model.CommentStatusRemoved:
		s.logger.Error("Comment is removed")
		return nil, network.NewForbiddenError(
			"Comment is removed",
			fmt.Sprintf("Comment with ID '%s' is removed. It cannot be replied to.", comment.CommentId),
			fmt.Errorf("comment %s is removed", comment.CommentId),
		)

	case model.CommentStatusArchived:
		s.logger.Error("Comment is archived")
		return nil, network.NewForbiddenError(
			"Comment is archived",
			fmt.Sprintf("Comment with ID '%s' is archived. It cannot be replied to.", comment.CommentId),
			fmt.Errorf("comment %s is archived", comment.CommentId),
		)
	}

	if commentModel.IsDeleted {
		s.logger.Error("Comment is deleted")
		return nil, network.NewForbiddenError(
			"Comment is deleted",
			fmt.Sprintf("Comment with ID '%s' is deleted. It cannot be replied to.", comment.CommentId),
			fmt.Errorf("comment %s is deleted", comment.CommentId),
		)
	}

	replyComment := model.NewComment(commentModel.PostId, userId, commentModel.CommunityId, comment.Reply, comment.CommentId)
	replyComment.AddDeviceInfo(comment.DeviceId, comment.DeviceType, comment.DeviceOS, comment.DeviceVersion)
	replyComment.AddLocationInfo(comment.Country, comment.City, comment.Latitude, comment.Longitude, comment.IpAddress, comment.TimeZone)
	replyComment.Path = fmt.Sprintf("%s.%s", commentModel.Path, commentModel.CommentId)
	replyComment.ParentId = commentModel.CommentId

	_, err = s.commentQueryBuilder.SingleQuery().InsertOne(replyComment)
	if err != nil {
		s.logger.Error("Failed to create post comment reply - %v", err)
		return nil, network.NewInternalServerError(
			"Failed to create comment reply",
			fmt.Sprintf("Failed to create comment reply - %s Context - [Query Failed]", err),
			network.DB_ERROR,
			err,
		)
	}

	// update the comment with the new reply
	update := bson.M{
		"$set": bson.M{
			"status":    model.CommentStatusActive,
			"updatedAt": replyComment.UpdatedAt,
		},
		"$inc": bson.M{
			"replyCount": 1,
		},
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(commentFilter, update, nil)
	if err != nil {
		s.logger.Error("Failed to update post comment with reply - %v", err)
		return nil, network.NewInternalServerError(
			"Failed to update comment with reply",
			fmt.Sprintf("Failed to update comment with reply - %s Context - [Query Failed]", err),
			network.DB_ERROR,
			err,
		)
	}

	return replyComment, nil
}

func (s *commentService) EditPostCommentReply(userId string, commentId string, comment *dto.EditCommentReplyRequest) (*model.Comment, network.ApiError) {
	filter := bson.M{"commentId": commentId}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(filter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return nil, network.NewNotFoundError(
			"Comment not found",
			fmt.Sprintf("Comment with ID '%s' not found, it may have been deleted or the ID is incorrect", commentId),
			nil,
		)
	}

	if commentModel.AuthorId != userId {
		s.logger.Error("User is not authorized to edit this comment")
		return nil, network.NewForbiddenError(
			"User is not authorized to edit this comment",
			fmt.Sprintf("User with ID '%s' is not authorized to edit comment with ID '%s', since it was created by user '%s'", userId, commentId, commentModel.AuthorId),
			fmt.Errorf("user %s is not authorized to edit comment %s", userId, commentId),
		)
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
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(filter, update, nil)
	if err != nil {
		s.logger.Error("Failed to update post comment reply - %v", err)
		return nil, network.NewInternalServerError(
			"Failed to update comment reply",
			fmt.Sprintf("Failed to update comment reply - %s Context - [Query Failed]", err),
			network.DB_ERROR,
			err,
		)
	}

	return commentModel, nil
}

func (s *commentService) DeletePostCommentReply(userId string, commentId string) network.ApiError {
	filter := bson.M{"commentId": commentId}
	commentModel, err := s.commentQueryBuilder.SingleQuery().FindOne(filter, nil)
	if err != nil {
		s.logger.Error("Failed to find comment - %v", err)
		return network.NewNotFoundError(
			"Comment not found",
			fmt.Sprintf("Comment with ID '%s' not found, it may have been deleted or the ID is incorrect", commentId),
			nil,
		)
	}
	if commentModel.AuthorId != userId {
		s.logger.Error("User is not authorized to delete this comment")
		return network.NewForbiddenError(
			"User is not authorized to delete this comment",
			fmt.Sprintf("User with ID '%s' is not authorized to delete comment with ID '%s', since it was created by user '%s'", userId, commentId, commentModel.AuthorId),
			fmt.Errorf("user %s is not authorized to delete comment %s", userId, commentId),
		)
	}
	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"commentId": commentId},
		bson.M{
			"$set": bson.M{
				"status":    model.CommentStatusDeleted,
				"isDeleted": true,
				"deletedAt": primitive.NewDateTimeFromTime(time.Now()),
				"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
			},
		},
		nil,
	)
	if err != nil {
		s.logger.Error("Failed to delete post comment reply - %v", err)
		return network.NewInternalServerError(
			"Failed to delete comment reply",
			fmt.Sprintf("Failed to delete comment reply - %s Context[ Query Failed : %v]", filter, err),
			network.DB_ERROR,
			err,
		)
	}

	_, err = s.commentQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"commentId": commentModel.ParentId},
		bson.M{
			"$set": bson.M{
				"status":    model.CommentStatusActive,
				"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
			},
			"$inc": bson.M{
				"replyCount": -1,
			},
		},
		nil,
	)
	if err != nil {
		s.logger.Error("Failed to update post comment reply - %v", err)
		return network.NewInternalServerError(
			"Failed to update comment reply",
			fmt.Sprintf("Failed to update comment reply - %s Context[ Query Failed : %v]", filter, err),
			network.DB_ERROR,
			err,
		)
	}
	return nil
}

func (s *commentService) LikePostComment(userId string, commentId string) (*bool, *int, network.ApiError) {
	err := s.toggleCommentInteraction(userId, commentId, model.CommentInteractionTypeLike)
	if err != nil {
		s.logger.Error("Failed to like post comment - %v", err)
		return nil, nil, err
	}

	commentSynergy, mongoErr := s.commentQueryBuilder.SingleQuery().FindOne(
		bson.M{"commentId": commentId},
		options.FindOne().SetProjection(bson.M{"synergy": -1}),
	)
	if mongoErr != nil {
		s.logger.Error("Failed to get comment synergy - %v", err)
		return nil, nil, network.NewInternalServerError(
			"Failed to get comment synergy",
			fmt.Sprintf("Failed to get comment synergy - %s Context - [Query Failed]", err),
			network.DB_ERROR,
			err,
		)
	}

	commentInteraction, mongoErr := s.commentInteractionQueryBuilder.SingleQuery().FindOne(
		bson.M{"commentId": commentId, "userId": userId},
		options.FindOne().SetProjection(bson.M{"interactionType": 1}),
	)
	if mongoErr != nil {
		if mongo.IsNoDocumentFoundError(mongoErr) {
			s.logger.Error("Comment interaction not found - %v", err)
			falseValue := false
			return &falseValue, &commentSynergy.Synergy, nil
		}
		s.logger.Error("Failed to get comment interaction - %v", err)
		return nil, nil, network.NewInternalServerError(
			"Failed to get comment interaction",
			fmt.Sprintf("Failed to get comment interaction - %s Context - [Query Failed]", err),
			network.DB_ERROR,
			err,
		)
	}
	var isLiked *bool
	if commentInteraction != nil {
		if commentInteraction.InteractionType == model.CommentInteractionTypeLike {
			trueValue := true
			isLiked = &trueValue
		} else {
			falseValue := false
			isLiked = &falseValue
		}
	} else {
		falseValue := false
		isLiked = &falseValue
	}
	return isLiked, &commentSynergy.Synergy, nil
}

func (s *commentService) DislikePostComment(userId string, commentId string) (*bool, *int, network.ApiError) {
	err := s.toggleCommentInteraction(userId, commentId, model.CommentInteractionTypeDislike)
	if err != nil {
		s.logger.Error("Failed to dislike post comment - %v", err)
		return nil, nil, err
	}

	commentSynergy, mongoErr := s.commentQueryBuilder.SingleQuery().FindOne(
		bson.M{"commentId": commentId},
		options.FindOne().SetProjection(bson.M{"synergy": -1}),
	)
	if mongoErr != nil {
		s.logger.Error("Failed to get comment synergy - %v", err)
		return nil, nil, network.NewInternalServerError(
			"Failed to get comment synergy",
			fmt.Sprintf("Failed to get comment synergy - %s Context - [Query Failed]", err),
			network.DB_ERROR,
			err,
		)
	}

	commentInteraction, mongoErr := s.commentInteractionQueryBuilder.SingleQuery().FindOne(
		bson.M{"commentId": commentId, "userId": userId},
		options.FindOne().SetProjection(bson.M{"interactionType": 1}),
	)
	if mongoErr != nil {
		if mongo.IsNoDocumentFoundError(mongoErr) {
			s.logger.Error("Comment interaction not found - %v", err)
			falseValue := false
			return &falseValue, &commentSynergy.Synergy, nil
		}
		s.logger.Error("Failed to get comment interaction - %v", err)
		return nil, nil, network.NewInternalServerError(
			"Failed to get comment interaction",
			fmt.Sprintf("Failed to get comment interaction - %s Context - [Query Failed]", err),
			network.DB_ERROR,
			err,
		)
	}

	var isDisliked *bool
	if commentInteraction != nil {
		if commentInteraction.InteractionType == model.CommentInteractionTypeDislike {
			trueValue := true
			isDisliked = &trueValue
		} else {
			falseValue := false
			isDisliked = &falseValue
		}
	} else {
		falseValue := false
		isDisliked = &falseValue
	}
	return isDisliked, &commentSynergy.Synergy, nil
}

func (s *commentService) toggleCommentInteraction(userId string, commentId string, interactionType model.CommentInteractionType) network.ApiError {
	action := "liking"
	if interactionType == model.CommentInteractionTypeDislike {
		action = "disliking"
	}
	s.logger.Info("%s comment with ID: %s by user: %s", action, commentId, userId)
	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)

	if err := tx.Start(); err != nil {
		s.logger.Error("Failed to start transaction: %v", err)
		return network.NewInternalServerError(
			"Failed to start transaction",
			fmt.Sprintf("Starting transaction failed - %s Context - [Transaction Failed]", err),
			network.DB_ERROR,
			err,
		)
	}

	var txErr error
	defer func() {
		if txErr != nil {
			if abortErr := tx.Abort(); abortErr != nil {
				s.logger.Error("Failed to abort transaction: %v", abortErr)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				s.logger.Error("Failed to commit transaction: %v", commitErr)
				txErr = commitErr
			}
		}
	}()

	result, txErr := tx.FindOne(
		model.CommentCollectionName,
		bson.M{"commentId": commentId, "status": model.CommentStatusActive},
	)
	if txErr != nil {
		s.logger.Error("Failed to get comment: %v", txErr)
		return network.NewInternalServerError(
			"Failed to get comment",
			fmt.Sprintf("Failed to get comment - %s Context - [Query Failed]", txErr),
			network.DB_ERROR,
			txErr,
		)
	}
	var comment *model.Comment
	result.Decode(&comment)
	if comment == nil {
		s.logger.Error("Comment not found")
		return network.NewNotFoundError(
			"Comment not found",
			fmt.Sprintf("Comment with ID '%s' not found, it may have been deleted or the ID is incorrect", commentId),
			nil,
		)
	}

	if result.IsNotFound() {
		s.logger.Error("Comment not found")
		return network.NewNotFoundError(
			"Comment not found",
			fmt.Sprintf("Comment with ID '%s' not found, it may have been deleted or the ID is incorrect", commentId),
			nil,
		)
	}

	commentResult, txErr := tx.FindMany(
		model.CommentInteractionCollectionName,
		bson.M{
			"commentId": commentId,
			"userId":    userId,
			"interactionType": bson.M{"$in": []model.CommentInteractionType{
				model.CommentInteractionTypeLike,
				model.CommentInteractionTypeDislike,
			}},
		},
	)
	if txErr != nil {
		s.logger.Error("Failed to get comment interactions: %v", txErr)
		return network.NewInternalServerError(
			"Failed to get comment interactions",
			fmt.Sprintf("Failed to get comment interactions - %s Context - [Query Failed]", txErr),
			network.DB_ERROR,
			txErr,
		)
	}
	existingInteractions := []*model.CommentInteraction{}
	if commentResult != nil {
		commentResult.All(&existingInteractions)
	}
	if existingInteractions == nil {
		existingInteractions = []*model.CommentInteraction{}
	}
	if commentResult.Err() != nil {
		if mongo.IsNoDocumentFoundError(commentResult.Err()) {
			s.logger.Error("Comment interaction not found - %v", commentResult.Err())
			return network.NewNotFoundError(
				"Comment interaction not found",
				fmt.Sprintf("Comment interaction with ID '%s' not found, it may have been deleted or the ID is incorrect", commentId),
				nil,
			)
		}
		s.logger.Error("Failed to get comment interactions: %v", commentResult.Err())
		return network.NewInternalServerError(
			"Failed to get comment interactions",
			fmt.Sprintf("Failed to get comment interactions - %s Context - [Query Failed]", commentResult.Err()),
			network.DB_ERROR,
			commentResult.Err(),
		)
	}

	synergyChange := 0
	needToInsert := true
	needToRemove := false
	removeID := ""

	if len(existingInteractions) == 0 {
		// First interaction
		if interactionType == model.CommentInteractionTypeLike {
			synergyChange = 1
		} else {
			synergyChange = -1
		}
	} else if len(existingInteractions) == 1 {
		existing := existingInteractions[0]
		needToRemove = true
		removeID = existing.Id.Hex()

		if existing.InteractionType == interactionType {
			// Toggle off the same interaction
			needToInsert = false
			if interactionType == model.CommentInteractionTypeLike {
				synergyChange = -1
			} else {
				synergyChange = 1
			}
		} else {
			// Switching between like and dislike
			if interactionType == model.CommentInteractionTypeLike {
				synergyChange = 2
			} else {
				synergyChange = -2
			}
		}
	} else {
		// Clean up duplicate interactions
		s.logger.Warn("Multiple interactions found for user %s on comment %s - cleaning up", userId, commentId)
		_, txErr = tx.DeleteMany(
			model.CommentInteractionCollectionName,
			bson.M{"commentId": commentId, "userId": userId},
		)
		if txErr != nil {
			s.logger.Error("Failed to clean up duplicate interactions: %v", txErr)
			return network.NewInternalServerError(
				"Failed to clean up duplicate interactions",
				fmt.Sprintf("Failed to clean up duplicate interactions - %s Context - [Query Failed]", txErr),
				network.DB_ERROR,
				txErr,
			)
		}

		if interactionType == model.CommentInteractionTypeLike {
			synergyChange = 1
		} else {
			synergyChange = -1
		}
	}

	// Update the comment synergy score
	updateResult := tx.FindOneAndUpdate(
		model.CommentCollectionName,
		bson.M{"commentId": commentId, "status": model.CommentStatusActive},
		bson.M{
			"$set": bson.M{"updatedAt": primitive.NewDateTimeFromTime(time.Now())},
			"$inc": bson.M{"synergy": synergyChange},
		},
	)

	if updateResult != nil {
		s.logger.Error("Failed to update comment synergy: %v", updateResult)
		txErr = updateResult
		return network.NewInternalServerError(
			"Failed to update comment",
			fmt.Sprintf("Failed to update comment - %s Context - [Query Failed]", updateResult),
			network.DB_ERROR,
			txErr,
		)
	}

	// Remove existing interaction if needed
	if needToRemove && removeID != "" {
		objID, _ := primitive.ObjectIDFromHex(removeID)
		_, txErr = tx.DeleteOne(
			model.CommentInteractionCollectionName,
			bson.M{"_id": objID},
		)
		if txErr != nil {
			s.logger.Error("Failed to remove existing interaction: %v", txErr)
			return network.NewInternalServerError(
				"Failed to update interaction",
				fmt.Sprintf("Failed to update interaction - %s Context - [Query Failed]", txErr),
				network.DB_ERROR,
				txErr,
			)
		}
	}

	// Insert new interaction if needed
	if needToInsert {
		commentInteraction := model.NewCommentInteraction(userId, commentId, interactionType)
		_, txErr := tx.InsertOne(model.CommentInteractionCollectionName, commentInteraction)
		if txErr != nil {
			if mongo.IsDuplicateKeyError(txErr) {
				s.logger.Warn("Comment interaction already exists (race condition): %v", txErr)
				txErr = nil
			} else {
				s.logger.Error("Failed to insert comment interaction: %v", txErr)
				return network.NewInternalServerError(
					"Failed to insert interaction",
					fmt.Sprintf("Failed to insert interaction - %s Context - [Query Failed]", txErr),
					network.DB_ERROR,
					txErr,
				)
			}
		}
	}

	s.logger.Info("Comment interaction updated successfully for comment ID: %s", commentId)
	return nil
}

func (s *commentService) GetUserComments(userId string, page int, limit int) ([]*model.PublicGetComment, network.ApiError) {
	s.logger.Debug("GetMyUserComments - userId: %s, page: %d, limit: %d", userId, page, limit)
	aggregate := s.commentAggregateBuilder.SingleAggregate()
	aggregate.Match(bson.M{"authorId": userId})
	aggregate.Sort(bson.D{{Key: "createdAt", Value: -1}, {Key: "synergy", Value: -1}})
	aggregate.Limit(int64(limit))
	aggregate.Skip(int64((page - 1) * limit))
	aggregate.Lookup("users", "authorId", "userId", "author")
	aggregate.Lookup("communities", "communityId", "communityId", "community")

	// Include the user's own interactions with their comments
	aggregate.Lookup(
		model.CommentInteractionCollectionName,
		"commentId",
		"commentId",
		"interactions",
	)

	aggregate.AddFields(bson.M{
		"author":    bson.M{"$arrayElemAt": bson.A{"$author", 0}},
		"community": bson.M{"$arrayElemAt": bson.A{"$community", 0}},
	})

	// Filter interactions for this specific user
	aggregate.AddFields(bson.M{
		"userInteractions": bson.M{
			"$filter": bson.M{
				"input": "$interactions",
				"as":    "interaction",
				"cond": bson.M{
					"$and": bson.A{
						bson.M{"$eq": bson.A{"$$interaction.userId", userId}},
						bson.M{"$in": bson.A{
							"$$interaction.interactionType",
							bson.A{model.CommentInteractionTypeLike, model.CommentInteractionTypeDislike},
						}},
					},
				},
			},
		},
	})

	aggregate.AddFields(bson.M{
		"userInteraction": bson.M{"$arrayElemAt": bson.A{"$userInteractions", 0}},
	}) // Project fields including isLiked/isDisliked flags
	aggregate.Project(bson.M{
		"id":       "$commentId",
		"postId":   1,
		"parentId": 1,
		"author": bson.M{
			"userId":     "$author.userId",
			"username":   "$author.username",
			"email":      "$author.email",
			"avatar":     "$author.avatar.profile.url",
			"background": "$author.avatar.background.url",
			"status":     "$author.status",
		},
		"community": bson.M{
			"id":          "$community.communityId",
			"name":        "$community.name",
			"description": "$community.description",
			"avatar":      "$community.media.avatar.url",
			"background":  "$community.media.background.url",
			"createdAt":   "$community.createdAt",
			"status":      "$community.status",
		},
		"isLiked": bson.M{
			"$cond": bson.A{
				bson.M{"$and": bson.A{
					bson.M{"$ifNull": bson.A{"$userInteraction", false}},
					bson.M{"$eq": bson.A{"$userInteraction.interactionType", model.CommentInteractionTypeLike}},
				}},
				true,
				false,
			},
		},
		"isDisliked": bson.M{
			"$cond": bson.A{
				bson.M{"$and": bson.A{
					bson.M{"$ifNull": bson.A{"$userInteraction", false}},
					bson.M{"$eq": bson.A{"$userInteraction.interactionType", model.CommentInteractionTypeDislike}},
				}},
				true,
				false,
			},
		},
		"content":          1,
		"formattedContent": 1,
		"status":           1,
		"synergy":          1,
		"replyCount":       1,
		"reactionCounts":   1,
		"level":            1,
		"isEdited":         1,
		"isPinned":         1,
		"isStickied":       1,
		"isLocked":         1,
		"isDeleted":        1,
		"isRemoved":        1,
		"hasMedia":         1,
		"mentions":         1,
		"path":             1,
		"createdAt":        1,
	})
	comments, err := aggregate.Exec()
	if err != nil {
		s.logger.Error("Failed to get my comments - %v", err)
		return nil, network.NewInternalServerError(
			"Failed to get comments",
			fmt.Sprintf("It seems the comments for user '%s' could not be retrieved - Aggregation failed. Please try again later. [Context: userId=%s]", userId, userId),
			network.DB_ERROR,
			err,
		)
	}

	if len(comments) == 0 {
		return []*model.PublicGetComment{}, nil
	} else {
		return comments, nil
	}
}
