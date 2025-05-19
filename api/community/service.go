package community

import (
	"errors"
	"fmt"
	"slices"
	"sync-backend/api/common/media"
	mediaMadels "sync-backend/api/common/media/model"
	"sync-backend/api/community/model"
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
	CreateCommunity(name string, description string, tags []string, avatarFilePath string, backgroundFilePath string, userId string) (*model.Community, network.ApiError)
	GetCommunityById(id string) (*model.Community, network.ApiError)
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
		return nil, network.NewInternalServerError("Error fetching community tags", network.DB_ERROR, err)
	}

	if len(communityTags) == 0 {
		s.logger.Error("No community tags found")
		return nil, network.NewInternalServerError("No community tags found", network.DB_ERROR, errors.New("no community tags found"))
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
			return nil, network.NewInternalServerError("Error checking for duplicate community", network.DB_ERROR, err)
		}
	}
	if duplicateCommunity != nil {
		if duplicateCommunity.Slug == community.Slug {
			s.logger.Error("Community with the same slug already exists")
			community.Slug = utils.GenerateUniqueSlug(community.Name)
		}
	}
	_, err = s.communityQueryBuilder.Query(s.Context()).InsertOne(community)
	if err != nil {
		s.logger.Error("Error inserting community: %v", err)
		return nil, network.NewInternalServerError("Error inserting community", network.DB_ERROR, err)
	}

	return community, nil
}

func (s *communityService) GetCommunityById(id string) (*model.Community, network.ApiError) {
	s.logger.Info("Fetching community with id: %s", id)
	filter := bson.M{"communityId": id}
	community, err := s.communityQueryBuilder.Query(s.Context()).FindOne(filter, nil)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		s.logger.Error("Error fetching community: %v", err)
		return nil, network.NewInternalServerError("Error fetching community", network.DB_ERROR, err)
	}
	if community == nil {
		s.logger.Error("Community not found")
		return nil, network.NewNotFoundError("Community not found", fmt.Errorf("community with id %s not found", id))
	}
	return community, nil
}

func (s *communityService) CheckUserInCommunity(userId string, communityId string) network.ApiError {
	s.logger.Info("Checking if user %s is in community %s", userId, communityId)
	community, err := s.communityQueryBuilder.Query(s.Context()).FindOne(bson.M{"communityId": communityId}, nil)
	if err != nil {
		s.logger.Error("Error fetching community: %v", err)
		return network.NewInternalServerError("Error fetching community", network.DB_ERROR, err)
	}

	if community == nil {
		s.logger.Error("Community not found")
		return network.NewNotFoundError("Community not found", errors.New("community not found"))
	}

	communityInteraction, err := s.communityInteractionQueryBuilder.Query(s.Context()).FindOne(bson.M{"communityId": communityId, "userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.logger.Error("Community interaction not found: %v", err)
			return network.NewNotFoundError("User is not a member of the community", err)
		}
		s.logger.Error("Error fetching community interaction: %v", err)
		return network.NewInternalServerError("Error fetching community interaction", network.DB_ERROR, err)
	}
	if communityInteraction.InteractionType == model.CommunityInteractionTypeJoin {
		s.logger.Info("User %s is a member of community %s", userId, communityId)
		return nil
	} else if communityInteraction.InteractionType == model.CommunityInteractionTypeLeave {
		s.logger.Info("User %s left the community %s", userId, communityId)
		return network.NewNotFoundError("User is not a member of the community", errors.New("user left the community"))
	} else {
		// This case should not happen, but just in case
		s.logger.Error("User is not a member of the community")
		return network.NewNotFoundError("User is not a member of the community", errors.New("user is not a member of the community"))
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
		return nil, network.NewInternalServerError("Error fetching communities", network.DB_ERROR, err)
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
		return nil, network.NewInternalServerError("Error executing community query", network.DB_ERROR, err)
	}
	aggregator.Close()
	return communityResults, nil
}

func (s *communityService) JoinCommunity(userId string, communityId string) network.ApiError {
	s.logger.Info("User %s is joining community %s", userId, communityId)

	// Start a transaction for consistent state
	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)
	defer tx.Abort()

	if err := tx.Start(); err != nil {
		s.logger.Error("Failed to start transaction: %v", err)
		return network.NewInternalServerError("Failed to start transaction", network.DB_ERROR, err)
	}

	communityCollection := tx.GetCollection(model.CommunityCollectionName)
	var community model.Community
	err := communityCollection.FindOne(
		tx.GetContext(),
		bson.M{"communityId": communityId},
	).Decode(&community)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.logger.Error("Community not found: %v", err)
			return network.NewNotFoundError("Community not found", err)
		}
		s.logger.Error("Error fetching community: %v", err)
		return network.NewInternalServerError("Error fetching community", network.DB_ERROR, err)
	}

	now := time.Now()
	ptNow := primitive.NewDateTimeFromTime(now)

	_, err = communityCollection.UpdateOne(
		tx.GetContext(),
		bson.M{"communityId": communityId},
		bson.M{
			"$inc": bson.M{
				"memberCount":              1,
				"stats.dailyActiveUsers":   1,
				"stats.weeklyActiveUsers":  1,
				"stats.monthlyActiveUsers": 1,
			},
			"$set": bson.M{
				"metadata.updatedAt": ptNow,
			},
		},
	)

	if err != nil {
		s.logger.Error("Error updating community: %v", err)
		return network.NewInternalServerError("Error updating community", network.DB_ERROR, err)
	}

	communityInteraction := model.NewCommunityInteraction(userId, communityId, model.CommunityInteractionTypeJoin)
	communityInteractionCollection := tx.GetCollection(model.CommunityInteractionsCollectionName)

	_, err = communityInteractionCollection.InsertOne(tx.GetContext(), communityInteraction)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			s.logger.Warn("Community interaction already exists (race condition): %v", err)
		} else {
			s.logger.Error("Failed to insert community interaction: %v", err)
			return network.NewInternalServerError("Failed to insert interaction", network.DB_ERROR, err)
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError("Failed to commit transaction", network.DB_ERROR, err)
	}

	s.logger.Info("User %s successfully joined community %s", userId, communityId)
	return nil
}

func (s *communityService) LeaveCommunity(userId string, communityId string) network.ApiError {
	s.logger.Info("User %s is leaving community %s", userId, communityId)

	// Start a transaction for consistent state
	tx := s.transaction.GetTransaction(mongo.DefaultShortTransactionTimeout)
	defer tx.Abort()

	if err := tx.Start(); err != nil {
		s.logger.Error("Failed to start transaction: %v", err)
		return network.NewInternalServerError("Failed to start transaction", network.DB_ERROR, err)
	}

	communityCollection := tx.GetCollection(model.CommunityCollectionName)
	var community model.Community
	err := communityCollection.FindOne(
		tx.GetContext(),
		bson.M{"communityId": communityId},
	).Decode(&community)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.logger.Error("Community not found: %v", err)
			return network.NewNotFoundError("Community not found", err)
		}
		s.logger.Error("Error fetching community: %v", err)
		return network.NewInternalServerError("Error fetching community", network.DB_ERROR, err)
	}

	// Check if user is the owner - owners cannot leave their community
	if community.OwnerId == userId {
		s.logger.Error("Owner cannot leave their community")
		return network.NewBadRequestError("Owner cannot leave their community", nil)
	}

	now := time.Now()
	ptNow := primitive.NewDateTimeFromTime(now)

	_, err = communityCollection.UpdateOne(
		tx.GetContext(),
		bson.M{"communityId": communityId},
		bson.M{
			"$inc": bson.M{
				"memberCount":              -1,
				"stats.dailyActiveUsers":   -1,
				"stats.weeklyActiveUsers":  -1,
				"stats.monthlyActiveUsers": -1,
			},
			"$set": bson.M{
				"metadata.updatedAt": ptNow,
			},
		},
	)

	if err != nil {
		s.logger.Error("Error updating community: %v", err)
		return network.NewInternalServerError("Error updating community", network.DB_ERROR, err)
	}

	communityInteraction := model.NewCommunityInteraction(userId, communityId, model.CommunityInteractionTypeLeave)
	communityInteractionCollection := tx.GetCollection(model.CommunityInteractionsCollectionName)

	_, err = communityInteractionCollection.InsertOne(tx.GetContext(), communityInteraction)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			s.logger.Warn("Community interaction already exists (race condition): %v", err)
		} else {
			s.logger.Error("Failed to insert community interaction: %v", err)
			return network.NewInternalServerError("Failed to insert interaction", network.DB_ERROR, err)
		}
	}

	// Also remove the user from moderators list if they are a moderator
	if slices.Contains(community.Moderators, userId) {
		_, err = communityCollection.UpdateOne(
			tx.GetContext(),
			bson.M{"communityId": communityId},
			bson.M{
				"$pull": bson.M{"moderators": userId},
			},
		)

		if err != nil {
			s.logger.Error("Error updating community moderators: %v", err)
			return network.NewInternalServerError("Error updating community moderators", network.DB_ERROR, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction: %v", err)
		return network.NewInternalServerError("Failed to commit transaction", network.DB_ERROR, err)
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
		return nil, network.NewInternalServerError("Error searching communities", network.DB_ERROR, err)
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
		return nil, network.NewInternalServerError("Error searching communities", network.DB_ERROR, err)
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
		return nil, network.NewInternalServerError("Error fetching trending communities", network.DB_ERROR, err)
	}

	aggregator.Close()
	return communitiesResults, nil
}
