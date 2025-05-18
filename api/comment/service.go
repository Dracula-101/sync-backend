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
	GetPostComments(userId string, postId string, page int, limit int) ([]*model.PublicComment, network.ApiError)
	GetPostCommentReplies(userId string, postId string, parentId string, page int, limit int) ([]*model.PublicComment, network.ApiError)

	CreatePostCommentReply(userId string, comment *dto.CreateCommentReplyRequest) (*model.Comment, network.ApiError)
	EditPostCommentReply(userId string, commentId string, comment *dto.EditCommentReplyRequest) (*model.Comment, network.ApiError)
	DeletePostCommentReply(userId string, commentId string) network.ApiError

	LikePostComment(userId string, commentId string) (*bool, *int, network.ApiError)
	DislikePostComment(userId string, commentId string) (*bool, *int, network.ApiError)

	GetUserComments(userId string, page int, limit int) ([]*model.PublicComment, network.ApiError)
}

type commentService struct {
	network.BaseService
	logger                         utils.AppLogger
	commentQueryBuilder            mongo.QueryBuilder[model.Comment]
	commentInteractionQueryBuilder mongo.QueryBuilder[model.CommentInteraction]
	postQueryBuilder               mongo.QueryBuilder[post.Post]
	communityQueryBuilder          mongo.QueryBuilder[community.Community]
	commentAggregateBuilder        mongo.AggregateBuilder[model.Comment, model.PublicComment]
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
		commentAggregateBuilder:        mongo.NewAggregateBuilder[model.Comment, model.PublicComment](db, model.CommentCollectionName),
		transaction:                    mongo.NewTransactionBuilder(db),
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

func (s *commentService) GetPostComments(userId string, postId string, page int, limit int) ([]*model.PublicComment, network.ApiError) {
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
			"avatar":     "$author.avatar.profilePic.url",
			"background": "$author.avatar.background.url",
			"status":     "$author.status",
		},
		"community": bson.M{
			"id":          "$community.communityId",
			"name":        "$community.name",
			"description": "$community.description",
			"avatar":      "$community.media.avatar.url",
			"background":  "$community.media.background.url",
			"createdAt":   "$community.metadata.createdAt",
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
		return nil, network.NewInternalServerError("Failed to get comments", network.DB_ERROR, err)
	}
	if len(comments) == 0 {
		return []*model.PublicComment{}, nil
	} else {
		return comments, nil
	}
}

func (s *commentService) GetPostCommentReplies(userId string, postId string, parentId string, page int, limit int) ([]*model.PublicComment, network.ApiError) {
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
			"avatar":     "$author.avatar.profilePic.url",
			"background": "$author.avatar.background.url",
			"status":     "$author.status",
		},
		"community": bson.M{
			"id":          "$community.communityId",
			"name":        "$community.name",
			"description": "$community.description",
			"avatar":      "$community.media.avatar.url",
			"background":  "$community.media.background.url",
			"createdAt":   "$community.metadata.createdAt",
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
		return nil, network.NewInternalServerError("Failed to get comments", network.DB_ERROR, err)
	}
	if len(comments) == 0 {
		return []*model.PublicComment{}, nil
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

	switch commentModel.Status {
	case model.CommentStatusDeleted:
		s.logger.Error("Comment is deleted")
		return nil, network.NewForbiddenError("Comment is deleted", fmt.Errorf("comment %s is deleted", comment.CommentId))
	case model.CommentStatusHidden:
		s.logger.Error("Comment is hidden")
		return nil, network.NewForbiddenError("Comment is hidden", fmt.Errorf("comment %s is hidden", comment.CommentId))
	case model.CommentStatusRemoved:
		s.logger.Error("Comment is removed")
		return nil, network.NewForbiddenError("Comment is removed", fmt.Errorf("comment %s is removed", comment.CommentId))
	case model.CommentStatusArchived:
		s.logger.Error("Comment is archived")
		return nil, network.NewForbiddenError("Comment is archived", fmt.Errorf("comment %s is archived", comment.CommentId))
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
				"status":           model.CommentStatusActive,
				"updatedAt":        primitive.NewDateTimeFromTime(time.Now()),
				"metadata.version": commentModel.Metadata.Version + 1,
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
		return nil, nil, network.NewInternalServerError("Failed to get comment synergy", network.DB_ERROR, err)
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
		return nil, nil, network.NewInternalServerError("Failed to get comment interaction", network.DB_ERROR, err)
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
		return nil, nil, network.NewInternalServerError("Failed to get comment synergy", network.DB_ERROR, err)
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
		return nil, nil, network.NewInternalServerError("Failed to get comment interaction", network.DB_ERROR, err)
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
	defer tx.Abort()

	if err := tx.Start(); err != nil {
		s.logger.Error("Failed to start transaction: %v", err)
		return network.NewInternalServerError("Failed to start transaction", network.DB_ERROR, err)
	}

	// First, check if the comment exists and is active
	commentCollection := tx.GetCollection(model.CommentCollectionName)
	var commentDoc model.Comment
	err := commentCollection.FindOne(
		tx.GetContext(),
		bson.M{"commentId": commentId, "status": model.CommentStatusActive},
	).Decode(&commentDoc)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.logger.Error("Comment not found or not active: %v", err)
			return network.NewNotFoundError("Comment not found or not active", err)
		}
		s.logger.Error("Failed to get comment: %v", err)
		return network.NewInternalServerError("Failed to get comment", network.DB_ERROR, err)
	}

	// Check for existing interactions by this user on this comment
	commentInteractionCollection := tx.GetCollection(model.CommentInteractionCollectionName)
	cursor, err := commentInteractionCollection.Find(
		tx.GetContext(),
		bson.M{
			"commentId": commentId,
			"userId":    userId,
			"interactionType": bson.M{"$in": []model.CommentInteractionType{
				model.CommentInteractionTypeLike,
				model.CommentInteractionTypeDislike,
			}},
		},
	)
	if err != nil {
		s.logger.Error("Failed to get comment interactions: %v", err)
		return network.NewInternalServerError("Failed to get comment interactions", network.DB_ERROR, err)
	}

	var existingInteractions []model.CommentInteraction
	if err := cursor.All(tx.GetContext(), &existingInteractions); err != nil {
		s.logger.Error("Failed to decode comment interactions: %v", err)
		return network.NewInternalServerError("Failed to decode comment interactions", network.DB_ERROR, err)
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
		_, deleteErr := commentInteractionCollection.DeleteMany(
			tx.GetContext(),
			bson.M{"commentId": commentId, "userId": userId},
		)
		if deleteErr != nil {
			s.logger.Error("Failed to clean up duplicate interactions: %v", deleteErr)
			return network.NewInternalServerError("Failed to clean up interactions", network.DB_ERROR, deleteErr)
		}

		if interactionType == model.CommentInteractionTypeLike {
			synergyChange = 1
		} else {
			synergyChange = -1
		}
	}

	// Update the comment synergy score
	updateResult := commentCollection.FindOneAndUpdate(
		tx.GetContext(),
		bson.M{"commentId": commentId, "status": model.CommentStatusActive},
		bson.M{
			"$set": bson.M{"updatedAt": primitive.NewDateTimeFromTime(time.Now())},
			"$inc": bson.M{"synergy": synergyChange},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if updateResult.Err() != nil {
		s.logger.Error("Failed to update comment synergy: %v", updateResult.Err())
		return network.NewInternalServerError("Failed to update comment", network.DB_ERROR, updateResult.Err())
	}

	// Remove existing interaction if needed
	if needToRemove && removeID != "" {
		objID, _ := primitive.ObjectIDFromHex(removeID)
		_, deleteErr := commentInteractionCollection.DeleteOne(
			tx.GetContext(),
			bson.M{"_id": objID},
		)
		if deleteErr != nil {
			s.logger.Error("Failed to remove existing interaction: %v", deleteErr)
			return network.NewInternalServerError("Failed to update interaction", network.DB_ERROR, deleteErr)
		}
	}

	// Insert new interaction if needed
	if needToInsert {
		commentInteraction := model.NewCommentInteraction(userId, commentId, interactionType)
		_, insertErr := commentInteractionCollection.InsertOne(tx.GetContext(), commentInteraction)
		if insertErr != nil {
			if mongo.IsDuplicateKeyError(insertErr) {
				s.logger.Warn("Comment interaction already exists (race condition): %v", insertErr)
			} else {
				s.logger.Error("Failed to insert comment interaction: %v", insertErr)
				return network.NewInternalServerError("Failed to insert interaction", network.DB_ERROR, insertErr)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError("Failed to commit transaction", network.DB_ERROR, err)
	}

	s.logger.Info("Comment interaction updated successfully for comment ID: %s", commentId)
	return nil
}

func (s *commentService) GetUserComments(userId string, page int, limit int) ([]*model.PublicComment, network.ApiError) {
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
			"avatar":     "$author.avatar.profilePic.url",
			"background": "$author.avatar.background.url",
			"status":     "$author.status",
		},
		"community": bson.M{
			"id":          "$community.communityId",
			"name":        "$community.name",
			"description": "$community.description",
			"avatar":      "$community.media.avatar.url",
			"background":  "$community.media.background.url",
			"createdAt":   "$community.metadata.createdAt",
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
		return nil, network.NewInternalServerError("Failed to get comments", network.DB_ERROR, err)
	}

	if len(comments) == 0 {
		return []*model.PublicComment{}, nil
	} else {
		return comments, nil
	}
}
