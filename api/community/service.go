package community

import (
	"errors"
	"fmt"
	"slices"
	"sync-backend/api/common/media"
	mediaMadels "sync-backend/api/common/media/model"
	"sync-backend/api/community/model"
	postModel "sync-backend/api/post/model"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommunityService interface {
	/* COMMUNITY CRUD */
	GetCommunityById(id string) (*model.PublicGetCommunity, network.ApiError)
	CreateCommunity(name string, description string, tags []string, avatarFilePath string, backgroundFilePath string, userId string) (*model.Community, network.ApiError)
	UpdateCommunity(id string, description string, avatarFilePath string, backgroundFilePath string, userId string) (*model.Community, network.ApiError)
	DeleteCommunity(id string, userId string) network.ApiError

	CheckUserInCommunity(userId string, communityId string) network.ApiError
	GetCommunities(userId string, page int, limit int) ([]*model.Community, network.ApiError)

	/* USER COMMUNITY INTERACTIONS */
	JoinCommunity(userId string, communityId string) network.ApiError
	LeaveCommunity(userId string, communityId string) network.ApiError

	/* COMMUNITY SEARCH */
	SearchCommunities(query string, page int, limit int, showPrivate bool) ([]*model.CommunitySearchResult, network.ApiError)
	AutocompleteCommunities(query string, page int, limit int, showPrivate bool) ([]*model.CommunityAutocomplete, network.ApiError)
	GetTrendingCommunities(page int, limit int) ([]*model.CommunitySearchResult, network.ApiError)
}

type communityService struct {
	network.BaseService
	mediaService                     media.MediaService
	logger                           utils.AppLogger
	communityQueryBuilder            mongo.QueryBuilder[model.Community]
	communityAggregateBuilder        mongo.AggregateBuilder[model.Community, model.Community]
	communityInteractionQueryBuilder mongo.QueryBuilder[model.CommunityInteraction]
	communityTagQueryBuilder         mongo.QueryBuilder[model.CommunityTag]
	getCommunityByIdPipeline         mongo.AggregateBuilder[model.Community, model.PublicGetCommunity]
	communitySearchPipeline          mongo.AggregateBuilder[model.Community, model.CommunitySearchResult]
	communityAutocompletePipeline    mongo.AggregateBuilder[model.Community, model.CommunityAutocomplete]
	transaction                      mongo.TransactionBuilder
}

func NewCommunityService(db mongo.Database, mediaService media.MediaService) CommunityService {
	return &communityService{
		mediaService:                     mediaService,
		BaseService:                      network.NewBaseService(),
		logger:                           utils.NewServiceLogger("CommunityService"),
		communityQueryBuilder:            mongo.NewQueryBuilder[model.Community](db, model.CommunityCollectionName),
		communityAggregateBuilder:        mongo.NewAggregateBuilder[model.Community, model.Community](db, model.CommunityCollectionName),
		communityInteractionQueryBuilder: mongo.NewQueryBuilder[model.CommunityInteraction](db, model.CommunityInteractionsCollectionName),
		communityTagQueryBuilder:         mongo.NewQueryBuilder[model.CommunityTag](db, model.CommunityTagCollectionName),
		getCommunityByIdPipeline:         mongo.NewAggregateBuilder[model.Community, model.PublicGetCommunity](db, model.CommunityCollectionName),
		communitySearchPipeline:          mongo.NewAggregateBuilder[model.Community, model.CommunitySearchResult](db, model.CommunityCollectionName),
		communityAutocompletePipeline:    mongo.NewAggregateBuilder[model.Community, model.CommunityAutocomplete](db, model.CommunityCollectionName),
		transaction:                      mongo.NewTransactionBuilder(db),
	}
}

func (s *communityService) CreateCommunity(name string, description string, tags []string, avatarfilePath string, backgroundFilePath string, userId string) (*model.Community, network.ApiError) {
	s.logger.Info("Creating community with name: %s", name)
	// get all community tags with the given tags
	filter := bson.M{"tag_id": bson.M{"$in": tags}}
	communityTags, err := s.communityTagQueryBuilder.Query(s.Context()).FindAll(filter, nil)
	if err != nil {
		s.logger.Error("Error fetching community tags: %v", err)
		return nil, NewDBError("fetching community tags", err.Error())
	}

	if len(communityTags) == 0 {
		s.logger.Error("No community tags found")
		return nil, NewDBError("fetching community tags", "no community tags found")
	}
	s.logger.Info("Community tags found: %v", communityTags)
	convertedTags := make([]model.CommunityTagInfo, len(communityTags))
	for i, tag := range communityTags {
		convertedTags[i] = tag.ToCommunityTagInfo()
	}

	var avatarPhoto mediaMadels.MediaInfo
	if avatarfilePath != "" {
		avatarPhoto, err = s.mediaService.UploadMedia(avatarfilePath, userId+"_avatar", "community")
		if err != nil {
			s.logger.Error("Error uploading media: %v", err)
		}
	} else {
		avatarPhoto = mediaMadels.MediaInfo{
			Id:     "default-avatar-community",
			Url:    "https:placehold.co/200x200.png",
			Width:  200,
			Height: 200,
		}
	}

	var backgroundPhoto mediaMadels.MediaInfo
	if backgroundFilePath != "" {
		backgroundPhoto, err = s.mediaService.UploadMedia(backgroundFilePath, userId+"_background", "community")
		if err != nil {
			s.logger.Error("Error uploading media: %v", err)
		}
	} else {
		backgroundPhoto = mediaMadels.MediaInfo{
			Id:     "default-background-community",
			Url:    "https://placehold.co/1400x300.png",
			Width:  1400,
			Height: 300,
		}
	}
	avatarPhotoInfo := model.Image{
		ID:     avatarPhoto.Id,
		Url:    avatarPhoto.Url,
		Width:  avatarPhoto.Width,
		Height: avatarPhoto.Height,
	}
	backgroundPhotoInfo := model.Image{
		ID:     backgroundPhoto.Id,
		Url:    backgroundPhoto.Url,
		Width:  backgroundPhoto.Width,
		Height: backgroundPhoto.Height,
	}
	community := model.NewCommunity(model.NewCommunityArgs{
		Name:        name,
		Description: description,
		OwnerId:     userId,
		Avatar:      avatarPhotoInfo,
		Background:  backgroundPhotoInfo,
		Tags:        convertedTags,
	})

	//check for duplicate community slug
	duplicateFilter := bson.M{"slug": community.Slug}
	duplicateCommunity, err := s.communityQueryBuilder.Query(s.Context()).FindOne(duplicateFilter, nil)
	if err != nil {
		if !mongo.IsNoDocumentFoundError(err) {
			s.logger.Error("Error checking for duplicate community: %v", err)
			return nil, NewDBError("checking for duplicate community", err.Error())
		}
	}
	if duplicateCommunity != nil {
		if duplicateCommunity.Slug == community.Slug {
			s.logger.Error("Community with the same slug already exists")
			community.Slug = utils.GenerateUniqueSlug(community.Name)
			return nil, NewDuplicateCommunityError(community.Slug)
		}
	}
	_, err = s.communityQueryBuilder.Query(s.Context()).InsertOne(community)
	if err != nil {
		s.logger.Error("Error inserting community: %v", err)
		return nil, NewDBError("inserting community", err.Error())
	}

	return community, nil
}

func (s *communityService) UpdateCommunity(id string, description string, avatarFilePath string, backgroundFilePath string, userId string) (*model.Community, network.ApiError) {

	s.logger.Info("Updating community with id: %s", id)
	filter := bson.M{"communityId": id}
	community, err := s.communityQueryBuilder.Query(s.Context()).FindOne(filter, nil)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		s.logger.Error("Error fetching community: %v", err)
		return nil, NewDBError("fetching community", err.Error())
	}
	if community == nil {
		s.logger.Error("Community not found")
		return nil, NewCommunityNotFoundError(id)
	}
	isOwner := false
	if community.OwnerId == userId {
		isOwner = true
	}
	isModerator := false
	if slices.Contains(community.Moderators, userId) {
		isModerator = true
	}
	if !isOwner && !isModerator {
		s.logger.Error("User is not the owner or moderator of the community")
		if !isOwner {
			return nil, NewNotAuthorizedError("user is not the owner of the community", userId, id)
		} else {
			return nil, NewNotAuthorizedError("user is not the moderator of the community", userId, id)
		}
	}
	var avatarPhoto mediaMadels.MediaInfo
	if avatarFilePath != "" {
		avatarPhoto, err = s.mediaService.UploadMedia(avatarFilePath, userId+"_avatar", "community")
		if err != nil {
			s.logger.Error("Error uploading media: %v", err)
			return nil, NewDBError("uploading media", err.Error())
		}
	}
	var backgroundPhoto mediaMadels.MediaInfo
	if backgroundFilePath != "" {
		backgroundPhoto, err = s.mediaService.UploadMedia(backgroundFilePath, userId+"_background", "community")
		if err != nil {
			s.logger.Error("Error uploading media: %v", err)
			return nil, NewDBError("uploading media", err.Error())
		}
	}

	avatarPhotoInfo := community.Media.Avatar
	if avatarFilePath != "" {
		avatarPhotoInfo = model.Image{
			ID:     avatarPhoto.Id,
			Url:    avatarPhoto.Url,
			Width:  avatarPhoto.Width,
			Height: avatarPhoto.Height,
		}
	}

	backgroundPhotoInfo := community.Media.Background
	if backgroundFilePath != "" {
		backgroundPhotoInfo = model.Image{
			ID:     backgroundPhoto.Id,
			Url:    backgroundPhoto.Url,
			Width:  backgroundPhoto.Width,
			Height: backgroundPhoto.Height,
		}
	}

	community.Description = description
	community.Media.Avatar = avatarPhotoInfo
	community.Media.Background = backgroundPhotoInfo
	community.Metadata.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	community.Metadata.UpdatedBy = userId

	update := bson.M{
		"$set": bson.M{
			"description":        community.Description,
			"media.avatar":       community.Media.Avatar,
			"media.background":   community.Media.Background,
			"metadata.updatedAt": primitive.NewDateTimeFromTime(time.Now()),
			"metadata.updatedBy": userId,
		},
	}
	_, err = s.communityQueryBuilder.Query(s.Context()).UpdateOne(filter, update, nil)
	if err != nil {
		s.logger.Error("Error updating community: %v", err)
		return nil, NewDBError("updating community", err.Error())
	}
	return community, nil
}

func (s *communityService) DeleteCommunity(id string, userId string) network.ApiError {
	s.logger.Info("Deleting community with id: %s", id)
	filter := bson.M{"communityId": id}
	community, err := s.communityQueryBuilder.Query(s.Context()).FindOne(filter, nil)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		s.logger.Error("Error fetching community: %v", err)
		return NewDBError("fetching community", err.Error())
	}
	if community == nil {
		s.logger.Error("Community not found")
		return NewCommunityNotFoundError(id)
	}

	if community.OwnerId != userId {
		s.logger.Error("User is not the owner of the community")
		return NewNotAuthorizedError("user is not the owner of the community", userId, id)
	}
	if community.Status != string(model.CommunityStatusActive) {
		s.logger.Error("Community is not active")
		return NewNotAuthorizedError("community is not active", userId, id)
	}

	// Start a transaction for consistent state
	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)
	err = tx.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		communityCollection := session.Collection(model.CommunityCollectionName)
		communityCollection.FindOneAndUpdate(
			bson.M{"communityId": id},
			bson.M{
				"$set": bson.M{
					"status":    model.CommunityStatusDeleted,
					"deletedAt": primitive.NewDateTimeFromTime(time.Now()),
				},
			},
		)
		if err != nil {
			if mongo.IsNoDocumentFoundError(err) {
				s.logger.Error("Community with id %s not found: %v", id, err)
				return network.NewNotFoundError("community not found", fmt.Sprintf("Community with ID '%s' not found. It may have been deleted or never existed. Context - [ No Data ] ", id), err)
			}
			s.logger.Error("Error updating community: %v", err)
			return network.NewInternalServerError("error updating community", fmt.Sprintf("Error updating community with ID '%s'. Context - [ Query Failed ] ", id), network.DB_ERROR, err)
		}

		// Delete all community interactions
		communityInteractionCollection := session.Collection(model.CommunityInteractionsCollectionName)
		_, err = communityInteractionCollection.UpdateMany(
			bson.M{"communityId": id},
			bson.M{
				"$set": bson.M{
					"deletedAt": primitive.NewDateTimeFromTime(time.Now()),
				},
			},
		)
		if err != nil {
			s.logger.Error("Error deleting community interactions: %v", err)
			return network.NewInternalServerError("error deleting community interactions", fmt.Sprintf("Error deleting community interactions for community %s. Context - [ Query Failed ] ", id), network.DB_ERROR, err)
		}

		// Delete all post in the community
		postCollection := session.Collection(postModel.PostCollectionName)
		_, err = postCollection.UpdateMany(
			bson.M{"communityId": id},
			bson.M{
				"$set": bson.M{
					"status":    postModel.PostStatusDeleted,
					"deletedAt": primitive.NewDateTimeFromTime(time.Now()),
				},
			},
		)

		// Delete all post interactions in the community
		postInteractionCollection := session.Collection(postModel.PostInteractionCollectionName)
		_, err = postInteractionCollection.UpdateMany(
			bson.M{"communityId": id},
			bson.M{
				"$set": bson.M{
					"status":    postModel.PostStatusDeleted,
					"deletedAt": primitive.NewDateTimeFromTime(time.Now()),
				},
			},
		)
		if err != nil {
			s.logger.Error("Error deleting post interactions: %v", err)
			return network.NewInternalServerError("error deleting post interactions", fmt.Sprintf("Error deleting post interactions for community %s. Context - [ Query Failed ] ", id), network.DB_ERROR, err)
		}

		return nil
	})
	if err != nil {
		if network.IsApiError(err) {
			s.logger.Error("Failed to delete community: %v", err)
			return network.AsApiError(err)
		}
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError("failed to commit transaction", fmt.Sprintf("Failed to commit transaction for community %s. Context - [ Transaction Failed ] ", id), network.DB_ERROR, err)
	}
	s.logger.Info("Community with id %s deleted successfully", id)
	return nil
}

func (s *communityService) GetCommunityById(id string) (*model.PublicGetCommunity, network.ApiError) {
	s.logger.Info("Fetching community with id: %s", id)
	getCommunityByIdPipeline := s.getCommunityByIdPipeline.SingleAggregate()
	getCommunityByIdPipeline.AllowDiskUse(true)
	getCommunityByIdPipeline.Match(bson.M{"communityId": id, "status": model.CommunityStatusActive})

	// Lookup community interactions to check if user has joined
	getCommunityByIdPipeline.Lookup(model.CommunityInteractionsCollectionName, "communityId", "communityId", "interactions")

	// Lookup moderator details from users collection
	getCommunityByIdPipeline.AddFields(bson.M{"moderatorIds": "$moderators"})

	// Lookup user details for moderators
	getCommunityByIdPipeline.Lookup("users", "moderators", "userId", "moderatorUsers")

	// Add fields to process the interaction data
	getCommunityByIdPipeline.AddFields(bson.M{
		"interactions": bson.M{
			"$filter": bson.M{
				"input": "$interactions",
				"as":    "interaction",
				"cond": bson.M{
					"$and": []bson.M{
						{"$eq": []string{"$$interaction.interactionType", string(model.CommunityInteractionTypeJoin)}},
						{"$eq": []string{"$$interaction.status", string(model.CommunityInteractionStatusActive)}},
					},
				},
			},
		},
		"moderators": bson.M{
			"$map": bson.M{
				"input": "$moderatorUsers",
				"as":    "moderator",
				"in": bson.M{
					"userId":     "$$moderator.userId",
					"username":   "$$moderator.username",
					"email":      "$$moderator.email",
					"avatar":     "$$moderator.avatar.profile.url",
					"background": "$$moderator.avatar.background.url",
					"status":     "$$moderator.status",
				},
			},
		},
	})

	// Add isJoined field
	getCommunityByIdPipeline.AddFields(bson.M{
		"isJoined": bson.M{
			"$cond": bson.M{
				"if":   bson.M{"$gt": []any{bson.M{"$size": "$interactions"}, 0}},
				"then": true,
				"else": false,
			},
		},
	})

	// Project the needed fields (include only the fields you want)
	getCommunityByIdPipeline.Project(bson.M{
		"communityId": 1,
		"slug":        1,
		"name":        1,
		"description": 1,
		"shortDesc":   1,
		"ownerId":     1,
		"isPrivate":   1,
		"memberCount": 1,
		"postCount":   1,
		"media":       1,
		"tags":        1,
		"rules":       1,
		"moderators":  1, // Now contains user details
		"stats":       1,
		"settings":    1,
		"isJoined":    1,
		"metadata":    1,
		"status":      1,
	})

	communityResults, err := getCommunityByIdPipeline.Exec()
	if err != nil {
		s.logger.Error("Error executing community query: %v", err)
		return nil, network.NewInternalServerError(
			"Error executing community query",
			fmt.Sprintf("Error executing community query for id %s. Context - [ Query Failed ] ", id),
			network.DB_ERROR,
			err,
		)
	}

	if len(communityResults) == 0 {
		s.logger.Error("Community not found")
		return nil, network.NewNotFoundError(
			"Community not found",
			fmt.Sprintf("Community with ID '%s' not found. It may have been deleted or never existed. Context - [ No Data ] ", id),
			errors.New("community not found"),
		)
	}

	getCommunityByIdPipeline.Close()
	return communityResults[0], nil
}

func (s *communityService) CheckUserInCommunity(userId string, communityId string) network.ApiError {
	s.logger.Info("Checking if user %s is in community %s", userId, communityId)
	community, err := s.communityQueryBuilder.Query(s.Context()).FindOne(bson.M{"communityId": communityId}, nil)
	if err != nil {
		s.logger.Error("Error fetching community: %v", err)
		return network.NewInternalServerError(
			"Error fetching community",
			fmt.Sprintf("Error fetching community with id %s. Context - [ Query Failed ] ", communityId),
			network.DB_ERROR,
			err,
		)
	}

	if community == nil {
		s.logger.Error("Community not found")
		return network.NewNotFoundError(
			"Community not found",
			fmt.Sprintf("It seems the community with ID '%s' does not exist. The community may have been deleted or never existed. Context - [ No Data ] ", communityId),
			errors.New("community not found"),
		)
	}

	communityInteraction, err := s.communityInteractionQueryBuilder.Query(s.Context()).FindOne(bson.M{"communityId": communityId, "userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.logger.Error("Community interaction not found: %v", err)
			return network.NewNotFoundError(
				"User is not a member of the community",
				fmt.Sprintf("User %s is not a member of community %s. Context - [ No Data ] ", userId, communityId),
				err,
			)
		}
		s.logger.Error("Error fetching community interaction: %v", err)
		return network.NewInternalServerError(
			"Error fetching community interaction",
			fmt.Sprintf("Error fetching community interaction for user %s in community %s. Context - [ Query Failed ] ", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}
	if communityInteraction.InteractionType == model.CommunityInteractionTypeJoin {
		s.logger.Info("User %s is a member of community %s", userId, communityId)
		return nil
	} else if communityInteraction.InteractionType == model.CommunityInteractionTypeLeave {
		s.logger.Info("User %s left the community %s", userId, communityId)
		return network.NewNotFoundError(
			"User is not a member of the community",
			fmt.Sprintf("User %s left the community %s. Context - [ No Data ] ", userId, communityId),
			errors.New("user is not a member of the community"),
		)
	} else {
		// This case should not happen, but just in case
		s.logger.Error("User is not a member of the community")
		return network.NewNotFoundError(
			"User is not a member of the community",
			fmt.Sprintf("User %s is not a member of community %s. Context - [ No Data ] ", userId, communityId),
			errors.New("user is not a member of the community"),
		)
	}
}

func (s *communityService) GetCommunities(userId string, page int, limit int) ([]*model.Community, network.ApiError) {
	s.logger.Info("Fetching communities for user %s, page: %d, limit: %d", userId, page, limit)

	communityInteractions, err := s.communityInteractionQueryBuilder.Query(s.Context()).FindPaginated(
		bson.M{"userId": userId, "interactionType": model.CommunityInteractionTypeJoin},
		int64(page),
		int64(limit),
		options.Find().SetSort(bson.M{"createdAt": -1}),
	)

	if err != nil {
		s.logger.Error("Error fetching communities: %v", err)
		return nil, network.NewInternalServerError(
			"Error fetching communities",
			fmt.Sprintf("Error fetching communities for user %s. Context - [ Query Failed ] ", userId),
			network.DB_ERROR,
			err,
		)
	}

	var communityIds []string
	for _, interaction := range communityInteractions {
		// check if the community already exists in the list
		if slices.Contains(communityIds, interaction.CommunityId) {
			continue
		}
		communityIds = append(communityIds, interaction.CommunityId)
	}

	aggregator := s.communityAggregateBuilder.SingleAggregate()
	aggregator.AllowDiskUse(true)
	aggregator.Match(bson.M{"communityId": bson.M{"$in": communityIds}, "status": model.CommunityStatusActive})
	aggregator.Skip(int64((page - 1) * limit))
	aggregator.Limit(int64(limit))
	communityResults, err := aggregator.Exec()
	if err != nil {
		s.logger.Error("Error executing community query: %v", err)
		return nil, network.NewInternalServerError(
			"Error executing community query",
			fmt.Sprintf("Error executing community query for user %s. Context - [ Query Failed ] ", userId),
			network.DB_ERROR,
			err,
		)
	}
	aggregator.Close()
	return communityResults, nil
}

func (s *communityService) JoinCommunity(userId string, communityId string) network.ApiError {
	s.logger.Info("User %s is joining community %s", userId, communityId)

	// Start a transaction for consistent state
	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)

	err := tx.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		communityCollection := session.Collection(model.CommunityCollectionName)
		now := time.Now()
		ptNow := primitive.NewDateTimeFromTime(now)

		mongoErr := communityCollection.FindOneAndUpdate(
			bson.M{"communityId": communityId, "status": model.CommunityStatusActive},
			bson.M{
				"$inc": bson.M{"memberCount": 1},
				"$set": bson.M{"metadata.updatedAt": ptNow},
			},
		)

		if mongoErr.Err() != nil {
			if mongo.IsNoDocumentFoundError(mongoErr.Err()) {
				s.logger.Error("Community with id %s not found: %v", communityId, mongoErr.Err())
				return network.NewNotFoundError(
					"Community not found",
					fmt.Sprintf("Community with ID '%s' not found. It may have been deleted or never existed. Context - [ No Data ] ", communityId),
					mongoErr.Err(),
				)
			}
			s.logger.Error("Error updating community: %v", mongoErr.Err())
			return network.NewInternalServerError(
				"Error updating community",
				fmt.Sprintf("Error updating community with ID '%s'. Context - [ Query Failed ] ", communityId),
				network.DB_ERROR,
				mongoErr.Err(),
			)
		}

		communityInteractionCollection := session.Collection(model.CommunityInteractionsCollectionName)
		communityInteraction := model.NewCommunityInteraction(userId, communityId, model.CommunityInteractionTypeJoin, model.CommunityInteractionStatusActive)
		_, insertErr := communityInteractionCollection.InsertOne(communityInteraction)
		if insertErr != nil {
			if mongo.IsDuplicateKeyError(insertErr) {
				s.logger.Warn("Community interaction already exists (race condition): %v", insertErr)
				return network.NewConflictError(
					"Community interaction already exists",
					fmt.Sprintf("User %s is already a member of community %s. Context - [ Duplicate Key ] ", userId, communityId),
					insertErr,
				)
			} else {
				s.logger.Error("Failed to insert community interaction: %v", insertErr)
				return network.NewInternalServerError(
					"Failed to insert community interaction",
					fmt.Sprintf("Failed to insert community interaction for user %s in community %s. Context - [ Query Failed ] ", userId, communityId),
					network.DB_ERROR,
					insertErr,
				)
			}
		}
		return nil
	})

	if err != nil {
		if network.IsApiError(err) {
			s.logger.Error("Failed to join community: %v", err)
			return network.AsApiError(err)
		}
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError(
			"Failed to commit transaction",
			fmt.Sprintf("Failed to commit transaction for user %s in community %s. Context - [ Transaction Failed ] ", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	s.logger.Info("User %s successfully joined community %s", userId, communityId)
	return nil
}

func (s *communityService) LeaveCommunity(userId string, communityId string) network.ApiError {
	s.logger.Info("User %s is leaving community %s", userId, communityId)

	// Start a transaction for consistent state
	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)

	err := tx.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		communityCollection := session.Collection(model.CommunityCollectionName)
		now := time.Now()
		ptNow := primitive.NewDateTimeFromTime(now)

		updateErr := communityCollection.FindOneAndUpdate(
			bson.M{"communityId": communityId, "status": model.CommunityStatusActive},
			bson.M{
				"$inc": bson.M{"memberCount": -1},
				"$set": bson.M{"metadata.updatedAt": ptNow},
			},
		)

		if updateErr.Err() != nil {
			if mongo.IsNoDocumentFoundError(updateErr.Err()) {
				s.logger.Error("Community with id %s not found: %v", communityId, updateErr.Err())
				return network.NewNotFoundError(
					"Community not found",
					fmt.Sprintf("Community with ID '%s' not found. It may have been deleted or never existed. Context - [ No Data ] ", communityId),
					updateErr.Err(),
				)
			}
			s.logger.Error("Error updating community: %v", updateErr.Err())
			return network.NewInternalServerError(
				"Error updating community",
				fmt.Sprintf("Error updating community with ID '%s'. Context - [ Query Failed ] ", communityId),
				network.DB_ERROR,
				updateErr.Err(),
			)
		}

		communityInteractionCollection := session.Collection(model.CommunityInteractionsCollectionName)
		insertErr := communityInteractionCollection.FindOneAndUpdate(
			bson.M{"userId": userId, "communityId": communityId, "interactionType": model.CommunityInteractionTypeJoin},
			bson.M{
				"$set": bson.M{
					"status":          model.CommunityInteractionStatusInactive,
					"interactionType": model.CommunityInteractionTypeLeave,
					"updatedAt":       ptNow,
					"deletedAt":       ptNow,
				},
			},
		)
		if insertErr.Err() != nil {
			if mongo.IsNoDocumentFoundError(insertErr.Err()) {
				// user hasnt joined the community yet
				s.logger.Error("Community interaction not found: %v", insertErr.Err())
				return network.NewNotFoundError(
					"Community interaction not found",
					fmt.Sprintf("Community interaction not found for user %s in community %s. Context - [ No Data ] ", userId, communityId),
					insertErr.Err(),
				)
			}
			s.logger.Error("Failed to update community interaction: %v", insertErr.Err())
			return network.NewInternalServerError(
				"Failed to update community interaction",
				fmt.Sprintf("Failed to update community interaction for user %s in community %s. Context - [ Query Failed ] ", userId, communityId),
				network.DB_ERROR,
				insertErr.Err(),
			)
		}

		// Also remove the user from moderators list if they are a moderator
		_, removeErr := communityCollection.UpdateOne(
			bson.M{"communityId": communityId},
			bson.M{
				"$pull": bson.M{"moderators": userId},
			},
		)
		if removeErr != nil {
			s.logger.Error("Failed to remove user from moderators list: %v", removeErr)
			return network.NewInternalServerError(
				"Failed to remove user from moderators list",
				fmt.Sprintf("Failed to remove user %s from moderators list in community %s. Context - [ Query Failed ] ", userId, communityId),
				network.DB_ERROR,
				removeErr,
			)
		}

		return nil
	})

	if err != nil {
		if network.IsApiError(err) {
			s.logger.Error("Failed to leave community: %v", err)
			return network.AsApiError(err)
		}
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError(
			"Failed to commit transaction",
			fmt.Sprintf("Failed to commit transaction for user %s in community %s. Context - [ Transaction Failed ] ", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	s.logger.Info("User %s successfully left community %s", userId, communityId)
	return nil
}

func (s *communityService) SearchCommunities(query string, page int, limit int, showPrivate bool) ([]*model.CommunitySearchResult, network.ApiError) {
	s.logger.Info("Searching communities with query: %s, page: %d, limit: %d", query, page, limit)

	aggregator := s.communitySearchPipeline.
		Aggregate(s.Context()).
		AllowDiskUse(true)

	matchStage := bson.M{
		"$and": []bson.M{
			{"status": model.CommunityStatusActive},
			{"$or": []bson.M{
				{"isPrivate": showPrivate},
				{"settings.showInDiscovery": true},
			}},
		},
	}

	if query != "" {
		searchQuery := bson.M{
			"index": "community_search",
			"compound": bson.M{
				"should": []bson.M{
					{
						"text": bson.M{
							"query":         query,
							"path":          "name",
							"score":         bson.M{"boost": bson.M{"value": 5}},
							"matchCriteria": "any",
						},
					},
					{
						"text": bson.M{
							"query": query,
							"path":  "slug",
							"score": bson.M{"boost": bson.M{"value": 3}},
						},
					},
					{
						"text": bson.M{
							"query": query,
							"path":  "shortDesc",
							"score": bson.M{"boost": bson.M{"value": 2}},
						},
					},
					{
						"text": bson.M{
							"query": query,
							"path":  "description",
							"score": bson.M{"boost": bson.M{"value": 1}},
						},
					},
					{
						"text": bson.M{
							"query": query,
							"path":  "tags.name",
							"score": bson.M{"boost": bson.M{"value": 4}},
						},
					},
				},
				"minimumShouldMatch": 1,
			},
			"highlight": bson.M{
				"path": []string{"name", "description", "shortDesc", "tags.name", "slug"},
			},
		}

		aggregator.Search("community_search", searchQuery)
	}

	aggregator.Match(matchStage)
	addFields := bson.M{
		"relevanceScore": bson.M{
			"$cond": bson.M{
				"if":   bson.M{"$gt": []any{"$memberCount", 0}},
				"then": bson.M{"$multiply": []any{bson.M{"$ifNull": []any{"$stats.popularityScore", 1}}, 1.5}},
				"else": 1,
			},
		},
		"score": bson.M{
			"$meta": "searchScore",
		},
		"matched": bson.M{
			"$map": bson.M{
				"input": bson.M{"$meta": "searchHighlights"},
				"as":    "highlight",
				"in":    "$$highlight.path",
			},
		},
	}
	aggregator.AddFields(addFields)

	projectStage := bson.M{
		"communityId":    1,
		"slug":           1,
		"name":           1,
		"description":    1,
		"shortDesc":      1,
		"ownerId":        1,
		"isPrivate":      1,
		"members":        1,
		"memberCount":    1,
		"postCount":      1,
		"stats":          1,
		"status":         1,
		"score":          1,
		"relevanceScore": 1,
		"matched":        1,
		"highlight": bson.M{
			"path": []string{"name", "description", "shortDesc", "tags.name", "slug"},
		},
	}
	aggregator.Project(projectStage)

	var sortStage bson.M
	if query != "" {
		sortStage = bson.M{
			"score":          -1,
			"relevanceScore": -1,
			"memberCount":    -1,
		}
	} else {
		sortStage = bson.M{
			"relevanceScore": -1,
			"memberCount":    -1,
			"createdAt":      -1,
		}
	}

	communitiesResults, err := aggregator.
		Sort(sortStage).
		Skip(int64((page - 1) * limit)).
		Limit(int64(limit)).
		Exec()

	if err != nil {
		s.logger.Error("Error executing community search: %v", err)
		return nil, network.NewInternalServerError(
			"Error executing community search",
			fmt.Sprintf("Error executing community search with query '%s'. Context - [ Query Failed ] ", query),
			network.DB_ERROR,
			err,
		)
	}

	aggregator.Close()
	return communitiesResults, nil
}

func (s *communityService) AutocompleteCommunities(query string, page int, limit int, showPrivate bool) ([]*model.CommunityAutocomplete, network.ApiError) {
	s.logger.Info("Autocomplete communities with query: %s, page: %d, limit: %d", query, page, limit)

	aggregator := s.communityAutocompletePipeline.
		Aggregate(s.Context()).
		AllowDiskUse(true)

	if query != "" {
		searchQuery := bson.M{
			"compound": bson.M{
				"should": []bson.M{
					{
						"autocomplete": bson.M{
							"query": query,
							"path":  "name",
							"score": bson.M{"boost": bson.M{"value": 5}},
						},
					},
					{
						"autocomplete": bson.M{
							"query": query,
							"path":  "tags.name",
							"score": bson.M{"boost": bson.M{"value": 3}},
						},
					},
					{
						"text": bson.M{
							"query": query,
							"path":  []string{"description", "shortDesc"},
							"fuzzy": bson.M{"maxEdits": 1},
							"score": bson.M{"boost": bson.M{"value": 2}},
						},
					},
				},
				"minimumShouldMatch": 1,
				"filter": []bson.M{
					{"text": bson.M{"query": "active", "path": "status"}},
					{"equals": bson.M{"path": "isPrivate", "value": showPrivate}},
				},
			},
			"highlight": bson.M{
				"path": []string{"name", "description", "shortDesc", "tags.name"},
			},
		}

		aggregator.Search("community_autocomplete", searchQuery)
	}

	matchStage := bson.M{
		"$and": []bson.M{
			{"status": "active"},
			{"$or": []bson.M{
				{"isPrivate": showPrivate},
				{"settings.showInDiscovery": true},
			}},
		},
	}
	aggregator.Match(matchStage)

	projectStage := bson.M{
		"communityId": 1,
		"slug":        1,
		"name":        1,
		"description": 1,
		"shortDesc":   1,
		"ownerId":     1,
		"isPrivate":   1,
		"members":     1,
		"memberCount": 1,
		"postCount":   1,
		"stats":       1,
		"status":      1,
		"score":       1,
	}
	aggregator.Project(projectStage)

	aggregator.AddFields(bson.M{
		"score": bson.M{
			"$meta": "searchScore",
		},
	})

	aggregator.Sort(bson.M{"score": -1, "memberCount": -1})
	aggregator.Skip(int64((page - 1) * limit))
	aggregator.Limit(int64(limit))

	communitiesResults, err := aggregator.Exec()
	if err != nil {
		s.logger.Error("Error executing community autocomplete: %v", err)
		return nil, network.NewInternalServerError(
			"Error executing community autocomplete",
			fmt.Sprintf("Error executing community autocomplete with query '%s'. Context - [ Query Failed ] ", query),
			network.DB_ERROR,
			err,
		)
	}

	aggregator.Close()
	return communitiesResults, nil
}

func (s *communityService) GetTrendingCommunities(page int, limit int) ([]*model.CommunitySearchResult, network.ApiError) {
	s.logger.Info("Fetching trending communities, page: %d, limit: %d", page, limit)

	aggregator := s.communitySearchPipeline.
		Aggregate(s.Context()).
		AllowDiskUse(true)

	matchStage := bson.M{
		"$and": []bson.M{
			{"status": "active"},
			{"$or": []bson.M{
				{"isPrivate": false},
				{"settings.showInDiscovery": true},
			}},
		},
	}
	aggregator.Match(matchStage)

	aggregator.AddFields(bson.M{
		"trendingScore": bson.M{
			"$add": []interface{}{
				bson.M{"$multiply": []interface{}{bson.M{"$ifNull": []interface{}{"$stats.engagementRate", 0}}, 3}},
				bson.M{"$multiply": []interface{}{bson.M{"$ifNull": []interface{}{"$stats.growthRate", 0}}, 2}},
				bson.M{"$multiply": []interface{}{bson.M{"$ifNull": []interface{}{"$stats.popularityScore", 0}}, 1.5}},
				bson.M{
					"$divide": []interface{}{
						1,
						bson.M{
							"$add": []interface{}{
								1,
								bson.M{
									"$divide": []interface{}{
										bson.M{"$subtract": []interface{}{"$$NOW", "$lastActivityAt"}},
										86400000, // MS in a day
									},
								},
							},
						},
					},
				},
			},
		},
	})

	projectStage := bson.M{
		"communityId":   1,
		"slug":          1,
		"name":          1,
		"description":   1,
		"shortDesc":     1,
		"ownerId":       1,
		"isPrivate":     1,
		"members":       1,
		"memberCount":   1,
		"postCount":     1,
		"stats":         1,
		"trendingScore": 1,
		"status":        1,
	}
	aggregator.Project(projectStage)

	sortStage := bson.M{
		"trendingScore": -1,
		"memberCount":   -1,
		"postCount":     -1,
	}

	communitiesResults, err := aggregator.
		Sort(sortStage).
		Skip(int64((page - 1) * limit)).
		Limit(int64(limit)).
		Exec()

	if err != nil {
		s.logger.Error("Error executing trending communities query: %v", err)
		return nil, network.NewInternalServerError(
			"Error executing trending communities query",
			"There was an error executing some operations. Context - [ Aggregation Failed ] ",
			network.DB_ERROR,
			err,
		)
	}

	aggregator.Close()
	return communitiesResults, nil
}
