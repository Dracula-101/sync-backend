package post

import (
	"fmt"
	"sync-backend/api/common/media"
	"sync-backend/api/community"
	"sync-backend/api/post/model"
	"sync-backend/api/user"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostService interface {
	CreatePost(title string, content string, tags []string, media []string, userId string, communityId string, postType model.PostType, isNSFW bool, isSpoiler bool) (*model.Post, error)
	GetPost(postId string, userId string) (*model.PublicPost, error)
	EditPost(userId string, postId string, title *string, content *string, postType model.PostType, isNSFW *bool, isSpoiler *bool) (*string, error)
	LikePost(userId string, postId string) (*bool, *int, error)
	DislikePost(userId string, postId string) (*bool, *int, error)
	SavePost(userId string, postId string) error
	SharePost(userId string, postId string) error

	GetPostsByUserId(userId string, page int, limit int) (posts []*model.Post, numOfPosts int, err error)
	GetPostsByCommunityId(communityId string, page int, limit int) (posts []*model.Post, numOfPosts int, err error)
}

type postService struct {
	network.BaseService
	mediaService                media.MediaService
	userService                 user.UserService
	logger                      utils.AppLogger
	communityService            community.CommunityService
	postQueryBuilder            mongo.QueryBuilder[model.Post]
	postInteractionQueryBuilder mongo.QueryBuilder[model.PostInteraction]
	getPostAggregateBuilder     mongo.AggregateBuilder[model.Post, model.PublicPost]
	transaction                 mongo.TransactionBuilder
}

func NewPostService(db mongo.Database, userService user.UserService, communityService community.CommunityService, mediaService media.MediaService) PostService {
	return &postService{
		BaseService:                 network.NewBaseService(),
		logger:                      utils.NewServiceLogger("PostService"),
		mediaService:                mediaService,
		userService:                 userService,
		communityService:            communityService,
		postQueryBuilder:            mongo.NewQueryBuilder[model.Post](db, model.PostCollectionName),
		postInteractionQueryBuilder: mongo.NewQueryBuilder[model.PostInteraction](db, model.PostInteractionCollectionName),
		getPostAggregateBuilder:     mongo.NewAggregateBuilder[model.Post, model.PublicPost](db, model.PostCollectionName),
		transaction:                 mongo.NewTransactionBuilder(db),
	}
}

func (s *postService) CreatePost(
	title string, content string, tags []string, media []string, userId string, communityId string, postType model.PostType, isNSFW bool, isSpoiler bool,
) (*model.Post, error) {
	s.logger.Info("Creating post with title: %s", title)
	var fileUrls []model.Media
	for _, file := range media {
		s.logger.Debug("File uploaded: %s", file)
		mediaInfo, err := s.mediaService.UploadMedia(file, userId+"_post", "post")
		if err != nil {
			s.logger.Error("Failed to upload media: %v", err)
			return nil, network.NewInternalServerError("Failed to upload media", network.MEDIA_ERROR, err)
		}
		fileUrls = append(fileUrls, model.Media{
			Id:        mediaInfo.Id,
			Type:      model.MediaType(mediaInfo.Type),
			Url:       mediaInfo.Url,
			Width:     mediaInfo.Width,
			Height:    mediaInfo.Height,
			FileSize:  mediaInfo.Size,
			CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		})
	}

	post := model.NewPost(userId, communityId, title, content, tags, fileUrls, postType, isNSFW, isSpoiler)

	if err := s.communityService.CheckUserInCommunity(userId, communityId); err != nil {
		s.logger.Error("User is not a member of the community: %v", err)
		return nil, network.NewForbiddenError("User is not a member of the community", err)
	}

	_, err := s.postQueryBuilder.SingleQuery().InsertOne(post)
	if err != nil {
		s.logger.Error("Failed to create post: %v", err)
		return nil, network.NewInternalServerError("Failed to create post", network.DB_ERROR, err)
	}
	s.logger.Info("Post created successfully with ID: %s", post.PostId)
	return post, nil
}

func (s *postService) GetPost(postId string, userId string) (*model.PublicPost, error) {
	s.logger.Info("Getting post with ID: %s", postId)
	// use aggregation to get the post with author and community details
	aggregate := s.getPostAggregateBuilder.SingleAggregate()
	aggregate.Match(bson.M{"postId": postId, "status": model.PostStatusActive})
	aggregate.Sort(bson.D{primitive.E{Key: "createdAt", Value: -1}, primitive.E{Key: "synergy", Value: -1}})
	aggregate.Lookup("users", "authorId", "userId", "author")
	aggregate.Lookup("communities", "communityId", "communityId", "community")
	aggregate.Lookup(model.PostInteractionCollectionName, "postId", "postId", "interactions")
	aggregate.AddFields(bson.M{
		"author":    bson.M{"$arrayElemAt": bson.A{"$author", 0}},
		"community": bson.M{"$arrayElemAt": bson.A{"$community", 0}},
		"userInteractions": bson.M{
			"$filter": bson.M{
				"input": "$interactions",
				"as":    "interaction",
				"cond": bson.M{
					"$and": bson.A{
						bson.M{"$eq": bson.A{"$$interaction.userId", userId}},
						bson.M{"$in": bson.A{
							"$$interaction.interactionType",
							bson.A{model.InteractionTypeLike, model.InteractionTypeDislike},
						}},
					},
				},
			},
		},
	}) // Calculate isLiked and isDisliked flags
	aggregate.AddFields(bson.M{
		"userInteraction": bson.M{"$arrayElemAt": bson.A{"$userInteractions", 0}},
	})
	aggregate.Project(bson.M{
		"id":      "$postId",
		"title":   1,
		"content": 1,
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
					bson.M{"$eq": bson.A{"$userInteraction.interactionType", model.InteractionTypeLike}},
				}},
				true,
				false,
			},
		},
		"isDisliked": bson.M{
			"$cond": bson.A{
				bson.M{"$and": bson.A{
					bson.M{"$ifNull": bson.A{"$userInteraction", false}},
					bson.M{"$eq": bson.A{"$userInteraction.interactionType", model.InteractionTypeDislike}},
				}},
				true,
				false,
			},
		},
		"type":         1,
		"status":       1,
		"media":        1,
		"tags":         1,
		"synergy":      1,
		"commentCount": 1,
		"viewCount":    1,
		"shareCount":   1,
		"saveCount":    1,
		"voters":       1,
		"isNSFW":       1,
		"isSpoiler":    1,
		"isStickied":   1,
		"isLocked":     1,
		"isArchived":   1,
		"createdAt":    1,
	})
	// execute the aggregation
	posts, err := aggregate.Exec()
	if err != nil {
		s.logger.Error("Failed to get post: %v", err)
		return nil, network.NewInternalServerError("Failed to get post", network.DB_ERROR, err)
	}
	if len(posts) == 0 {
		s.logger.Error("Post not found")
		return nil, network.NewNotFoundError("Post not found", fmt.Errorf("post with ID %s not found", postId))
	}
	s.logger.Info("Post retrieved successfully with ID: %s", postId)
	return posts[0], nil
}

func (s *postService) EditPost(userId string, postId string, title *string, content *string, postType model.PostType, isNSFW *bool, isSpoiler *bool) (newPostId *string, err error) {
	s.logger.Info("Editing post with ID: %s", postId)
	post, err := s.GetPost(postId, userId)
	if err != nil {
		return nil, err
	}
	if !post.IsActive() {
		s.logger.Error("Cannot edit inactive post with ID: %s", postId)
		return nil, network.NewForbiddenError("Cannot edit inactive post", fmt.Errorf("post with ID %s is not active", postId))
	}
	if post.Author.UserId != userId {
		s.logger.Error("User is not the author of the post: %s", postId)
		return nil, network.NewForbiddenError("User is not the author of the post", fmt.Errorf("user with ID %s is not the author of post %s", userId, postId))
	}

	filter := bson.M{"postId": postId, "authorId": userId}
	update := bson.M{}
	if title != nil {
		update["title"] = *title
	}
	if content != nil {
		update["content"] = *content
	}
	if postType != "" {
		update["type"] = postType
	}
	if isNSFW != nil {
		update["isNSFW"] = *isNSFW
	}
	if isSpoiler != nil {
		update["isSpoiler"] = *isSpoiler
	}
	update["updatedAt"] = primitive.NewDateTimeFromTime(time.Now())
	update["metadata"] = bson.M{
		"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
		"updatedBy": userId,
	}
	options := options.Update().SetUpsert(true)
	updatePost, err := s.postQueryBuilder.SingleQuery().UpdateOne(filter, bson.M{"$set": update}, options)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		s.logger.Error("Failed to edit post: %v", err)
		return nil, network.NewInternalServerError("Failed to edit post", network.DB_ERROR, err)
	}
	if updatePost == nil {
		s.logger.Error("Post not found")
		return nil, network.NewNotFoundError("Post not found", fmt.Errorf("post with ID %s not found", postId))
	}
	s.logger.Info("Post edited successfully with ID: %s -> New Id %s", postId, updatePost.UpsertedID)
	if updatePost.UpsertedID != nil {
		idStr := updatePost.UpsertedID.(string)
		return &idStr, nil
	}
	return nil, nil
}

func (s *postService) LikePost(userId string, postId string) (*bool, *int, error) {
	err := s.toggleInteraction(userId, postId, model.InteractionTypeLike)
	if err != nil {
		s.logger.Error("Failed to toggle like interaction: %v", err)
		return nil, nil, network.NewInternalServerError("Failed to toggle like interaction", network.DB_ERROR, err)
	}

	//get post synergy
	postSynergy, err := s.postQueryBuilder.SingleQuery().FindOne(
		bson.M{"postId": postId},
		options.FindOne().SetProjection(bson.M{"synergy": -1}),
	)
	if err != nil {
		s.logger.Error("Failed to get post synergy: %v", err)
		return nil, nil, network.NewInternalServerError("Failed to get post synergy", network.DB_ERROR, err)
	}
	//get post interaction
	postInteraction, err := s.postInteractionQueryBuilder.SingleQuery().FindOne(
		bson.M{"postId": postId, "userId": userId},
		options.FindOne().SetProjection(bson.M{"interactionType": 1}),
	)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			// user has unliked the post
			falseValue := false
			return &falseValue, &postSynergy.Synergy, nil
		}
		s.logger.Error("Failed to get post interaction: %v", err)
		return nil, nil, network.NewInternalServerError("Failed to get post interaction", network.DB_ERROR, err)
	}
	var isLiked *bool
	if postInteraction != nil {
		if postInteraction.InteractionType == model.InteractionTypeLike {
			trueValue := true
			isLiked = &trueValue
		} else {
			falseValue := false
			isLiked = &falseValue
		}
	} else {
		isLiked = nil
	}
	return isLiked, &postSynergy.Synergy, nil
}

func (s *postService) DislikePost(userId string, postId string) (*bool, *int, error) {
	err := s.toggleInteraction(userId, postId, model.InteractionTypeDislike)
	if err != nil {
		s.logger.Error("Failed to toggle dislike interaction: %v", err)
		return nil, nil, network.NewInternalServerError("Failed to toggle dislike interaction", network.DB_ERROR, err)
	}
	//get post synergy
	postSynergy, err := s.postQueryBuilder.SingleQuery().FindOne(
		bson.M{"postId": postId},
		options.FindOne().SetProjection(bson.M{"synergy": -1}),
	)
	if err != nil {
		s.logger.Error("Failed to get post synergy: %v", err)
		return nil, nil, network.NewInternalServerError("Failed to get post synergy", network.DB_ERROR, err)
	}
	//get post interaction
	postInteraction, err := s.postInteractionQueryBuilder.SingleQuery().FindOne(
		bson.M{"postId": postId, "userId": userId},
		options.FindOne().SetProjection(bson.M{"interactionType": 1}),
	)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			// user has unliked the post
			falseValue := false
			return &falseValue, &postSynergy.Synergy, nil
		}
		s.logger.Error("Failed to get post interaction: %v", err)
		return nil, nil, network.NewInternalServerError("Failed to get post interaction", network.DB_ERROR, err)
	}

	var isLiked *bool
	if postInteraction != nil {
		if postInteraction.InteractionType == model.InteractionTypeDislike {
			trueValue := true
			isLiked = &trueValue
		} else {
			falseValue := false
			isLiked = &falseValue
		}
	} else {
		isLiked = nil
	}
	return isLiked, &postSynergy.Synergy, nil
}

func (s *postService) toggleInteraction(userId string, postId string, interactionType model.InteractionType) error {
	action := "liking"
	if interactionType == model.InteractionTypeDislike {
		action = "disliking"
	}
	s.logger.Info("%s post with ID: %s by user: %s", action, postId, userId)

	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)
	if err := tx.Start(); err != nil {
		s.logger.Error("Failed to start transaction: %v", err)
		return network.NewInternalServerError("Failed to start transaction", network.DB_ERROR, err)
	}

	err := tx.PerformTransaction(func(session mongo.DatabaseSession) error {
		postInteractionCollection := session.Collection(model.PostInteractionCollectionName)
		cursor, err := postInteractionCollection.Find(
			bson.M{
				"postId": postId,
				"userId": userId,
				"interactionType": bson.M{"$in": []model.InteractionType{
					model.InteractionTypeLike,
					model.InteractionTypeDislike,
				}},
			},
		)
		if err != nil {
			s.logger.Error("Failed to get post interactions: %v", err)
			return network.NewInternalServerError("Failed to get post interactions", network.DB_ERROR, err)
		}

		var existingInteractions []model.PostInteraction
		if err := cursor.All(&existingInteractions); err != nil {
			s.logger.Error("Failed to decode post interactions: %v", err)
			return network.NewInternalServerError("Failed to decode post interactions", network.DB_ERROR, err)
		}

		synergyChange := 0
		needToInsert := true
		needToRemove := false
		removeID := ""

		if len(existingInteractions) == 0 {
			// First interaction
			if interactionType == model.InteractionTypeLike {
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
				if interactionType == model.InteractionTypeLike {
					synergyChange = -1
				} else {
					synergyChange = 1
				}
			} else {
				// Switching between like and dislike
				if interactionType == model.InteractionTypeLike {
					synergyChange = 2
				} else {
					synergyChange = -2
				}
			}
		} else {
			// Clean up duplicate interactions
			s.logger.Warn("Multiple interactions found for user %s on post %s - cleaning up", userId, postId)
			_, deleteErr := postInteractionCollection.DeleteMany(
				bson.M{"postId": postId, "userId": userId},
			)
			if deleteErr != nil {
				s.logger.Error("Failed to clean up duplicate interactions: %v", deleteErr)
				return network.NewInternalServerError("Failed to clean up interactions", network.DB_ERROR, deleteErr)
			}

			if interactionType == model.InteractionTypeLike {
				synergyChange = 1
			} else {
				synergyChange = -1
			}
		}

		postCollection := session.Collection(model.PostCollectionName)
		updateResult := postCollection.FindOneAndUpdate(
			bson.M{"postId": postId, "status": model.PostStatusActive},
			bson.M{
				"$set": bson.M{"lastActivityAt": primitive.NewDateTimeFromTime(time.Now())},
				"$inc": bson.M{"synergy": synergyChange},
			},
		)

		if updateResult.Err() != nil {
			s.logger.Error("Failed to update post synergy: %v", updateResult.Err())
			return network.NewInternalServerError("Failed to update post", network.DB_ERROR, updateResult.Err())
		}

		if needToRemove && removeID != "" {
			objID, _ := primitive.ObjectIDFromHex(removeID)
			_, deleteErr := postInteractionCollection.DeleteOne(
				bson.M{"_id": objID},
			)
			if deleteErr != nil {
				s.logger.Error("Failed to remove existing interaction: %v", deleteErr)
				return network.NewInternalServerError("Failed to update interaction", network.DB_ERROR, deleteErr)
			}
		}

		if needToInsert {
			postInteraction := model.NewPostInteraction(userId, postId, interactionType)
			_, insertErr := postInteractionCollection.InsertOne(postInteraction)
			if insertErr != nil {
				if mongo.IsDuplicateKeyError(insertErr) {
					s.logger.Warn("Post interaction already exists (race condition): %v", insertErr)
				} else {
					s.logger.Error("Failed to insert post interaction: %v", insertErr)
					return network.NewInternalServerError("Failed to insert interaction", network.DB_ERROR, insertErr)
				}
			}
		}
		return nil
	})

	if err != nil {
		if network.IsApiError(err) {
			s.logger.Error("Failed to toggle interaction: %v", err)
			return network.AsApiError(err)
		}
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError("Failed to commit transaction", network.DB_ERROR, err)
	}

	s.logger.Info("Post interaction updated successfully for post ID: %s", postId)
	return nil
}

func (s *postService) SavePost(userId string, postId string) error {
	s.logger.Info("Saving post with ID: %s", postId)
	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)

	if err := tx.Start(); err != nil {
		s.logger.Error("Failed to start transaction: %v", err)
		return network.NewInternalServerError("Failed to start transaction", network.DB_ERROR, err)
	}

	err := tx.PerformTransaction(func(session mongo.DatabaseSession) error {
		postInteractionCollection := session.Collection(model.PostInteractionCollectionName)
		// check if the user has already saved the post
		exists, mongoErr := postInteractionCollection.CountDocuments(
			bson.M{"postId": postId, "userId": userId, "interactionType": model.InteractionTypeSave},
		)
		if mongoErr != nil {
			s.logger.Error("Failed to check if post is already saved: %v", mongoErr)
			return network.NewInternalServerError("Failed to check if post is already saved", network.DB_ERROR, mongoErr)
		}
		if exists > 0 {
			s.logger.Warn("Post already saved by user: %s", postId)
			return nil
		}

		postCollection := session.Collection(model.PostCollectionName)
		// update the viewCount and lastActivityAt
		err := postCollection.FindOneAndUpdate(
			bson.M{"postId": postId, "status": model.PostStatusActive},
			bson.M{
				"$inc": bson.M{"saveCount": 1},
				"$set": bson.M{"lastActivityAt": primitive.NewDateTimeFromTime(time.Now())},
			},
		)
		if err.Err() != nil {
			s.logger.Error("Failed to save post: %v", err)
			return network.NewInternalServerError("Failed to save post", network.DB_ERROR, fmt.Errorf("failed to save post: %v", err))
		}
		// insert the interaction
		postInteraction := model.NewPostInteraction(userId, postId, model.InteractionTypeSave)
		_, insertErr := postInteractionCollection.InsertOne(postInteraction)
		if insertErr != nil {
			s.logger.Error("Failed to insert post interaction: %v", insertErr)
			return network.NewInternalServerError("Failed to insert post interaction", network.DB_ERROR, fmt.Errorf("failed to insert post interaction: %v", insertErr))
		}

		return nil
	})

	if err != nil {
		if network.IsApiError(err) {
			s.logger.Error("Failed to save post: %v", err)
			return err
		}
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError("Failed to commit transaction", network.DB_ERROR, err)
	}

	s.logger.Info("Post saved successfully with ID: %s", postId)
	return nil
}

func (s *postService) SharePost(userId string, postId string) error {
	s.logger.Info("Sharing post with ID: %s", postId)
	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)
	if err := tx.Start(); err != nil {
		s.logger.Error("Failed to start transaction: %v", err)
		return network.NewInternalServerError("Failed to start transaction", network.DB_ERROR, err)
	}

	err := tx.PerformTransaction(func(session mongo.DatabaseSession) error {
		postCollection := session.Collection(model.PostCollectionName)
		// update the viewCount and lastActivityAt
		result := postCollection.FindOneAndUpdate(
			bson.M{"postId": postId, "status": model.PostStatusActive},
			bson.M{
				"$inc": bson.M{"shareCount": 1},
				"$set": bson.M{"lastActivityAt": primitive.NewDateTimeFromTime(time.Now())},
			},
		)
		if err := result.Err(); err != nil {
			s.logger.Error("Failed to share post: %v", err)
			return network.NewInternalServerError("Failed to share post", network.DB_ERROR, fmt.Errorf("failed to share post: %v", err))
		}
		return nil
	})

	if err != nil {
		if network.IsApiError(err) {
			s.logger.Error("Failed to share post: %v", err)
			return err
		}
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError("Failed to commit transaction", network.DB_ERROR, err)
	}

	s.logger.Info("Post shared successfully with ID: %s", postId)
	return nil
}

func (s *postService) GetPostsByUserId(userId string, page int, limit int) (posts []*model.Post, numOfPosts int, err error) {
	s.logger.Info("Getting posts for user with ID: %s", userId)
	filter := bson.M{"authorId": userId}
	dbPosts, err := s.postQueryBuilder.SingleQuery().FilterPaginated(filter, int64(page), int64(limit), nil)
	if err != nil {
		s.logger.Error("Failed to get posts: %v", err)
		return nil, 0, network.NewInternalServerError("Failed to get posts", network.DB_ERROR, err)
	}
	nPosts, err := s.postQueryBuilder.SingleQuery().FilterCount(filter)
	if err != nil {
		s.logger.Error("Failed to count posts: %v", err)
		return nil, 0, network.NewInternalServerError("Failed to count posts", network.DB_ERROR, err)
	}
	s.logger.Info("Posts retrieved successfully for user with ID: %s", userId)
	return dbPosts, int(nPosts), nil
}

func (s *postService) GetPostsByCommunityId(communityId string, page int, limit int) (posts []*model.Post, numOfPosts int, err error) {
	s.logger.Info("Getting posts for community with ID: %s", communityId)
	filter := bson.M{"communityId": communityId}
	options := options.Find().SetSort(bson.D{primitive.E{Key: "createdAt", Value: -1}})
	dbPosts, err := s.postQueryBuilder.SingleQuery().FilterPaginated(filter, int64(page), int64(limit), options)
	if err != nil {
		s.logger.Error("Failed to get posts: %v", err)
		return nil, 0, network.NewInternalServerError("Failed to get posts", network.DB_ERROR, err)
	}
	nPosts, err := s.postQueryBuilder.SingleQuery().FilterCount(filter)
	if err != nil {
		s.logger.Error("Failed to count posts: %v", err)
		return nil, 0, network.NewInternalServerError("Failed to count posts", network.DB_ERROR, err)
	}
	s.logger.Info("Posts retrieved successfully for community with ID: %s", communityId)
	return dbPosts, int(nPosts), nil
}
