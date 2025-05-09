package community

import (
	"errors"
	"fmt"
	"sync-backend/api/community/model"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"slices"

	"go.mongodb.org/mongo-driver/bson"
)

type CommunityService interface {
	CreateCommunity(name string, description string, tags []string, avatarUrl string, backgroundUrl string, userId string) (*model.Community, network.ApiError)
	GetCommunityById(id string) (*model.Community, network.ApiError)
	CheckUserInCommunity(userId string, communityId string) network.ApiError
	SearchCommunities(query string, page int, limit int) ([]*model.CommunitySearchResult, network.ApiError)
}

type communityService struct {
	network.BaseService
	logger                    utils.AppLogger
	communityQueryBuilder     mongo.QueryBuilder[model.Community]
	communityTagQueryBuilder  mongo.QueryBuilder[model.CommunityTag]
	communityAggregateBuilder mongo.AggregateBuilder[model.Community, model.CommunitySearchResult]
}

func NewCommunityService(db mongo.Database) CommunityService {
	return &communityService{
		BaseService:               network.NewBaseService(),
		logger:                    utils.NewServiceLogger("CommunityService"),
		communityQueryBuilder:     mongo.NewQueryBuilder[model.Community](db, model.CommunityCollectionName),
		communityTagQueryBuilder:  mongo.NewQueryBuilder[model.CommunityTag](db, model.CommunityTagCollectionName),
		communityAggregateBuilder: mongo.NewAggregateBuilder[model.Community, model.CommunitySearchResult](db, model.CommunityCollectionName),
	}
}

func (s *communityService) CreateCommunity(name string, description string, tags []string, avatarUrl string, backgroundUrl string, userId string) (*model.Community, network.ApiError) {
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

	community := model.NewCommunity(model.NewCommunityArgs{
		Name:          name,
		Description:   description,
		OwnerId:       userId,
		AvatarUrl:     nil,
		BackgroundUrl: nil,
		Tags:          convertedTags,
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

	if slices.Contains(community.Members, userId) {
		return nil
	}
	s.logger.Error("User is not a member of the community")
	return network.NewForbiddenError("User is not a member of the community", errors.New("user is not a member of the community"))
}

func (s *communityService) SearchCommunities(query string, page int, limit int) ([]*model.CommunitySearchResult, network.ApiError) {
	s.logger.Info("Searching communities with query: %s, page: %d, limit: %d", query, page, limit)

	aggregator := s.communityAggregateBuilder.
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

	if len(communitiesResults) == 0 && query != "" {
		s.logger.Info("No communities found for query: %s, trying fuzzy search", query)

		backupAggregator := s.communityAggregateBuilder.
			Aggregate(s.Context()).
			AllowDiskUse(true)

		backupSearchQuery := bson.M{
			"index": "community_search",
			"text": bson.M{
				"query": query,
				"path":  []string{"name", "slug", "shortDesc", "tags.name"},
				"fuzzy": bson.M{"maxEdits": 2},
			},
			"highlight": bson.M{
				"path": []string{"name", "description", "shortDesc", "tags.name", "slug"},
			},
		}

		backupResults, backupErr := backupAggregator.
			Search("community_search", backupSearchQuery).
			Match(bson.M{"status": "active"}).
			Sort(bson.M{"score": -1, "memberCount": -1}).
			Skip(int64((page - 1) * limit)).
			Limit(int64(limit)).
			Exec()

		backupAggregator.Close()
		if backupErr == nil {
			communitiesResults = backupResults
		} else {
			s.logger.Error("Error executing backup community search: %v", backupErr)
			return nil, network.NewInternalServerError("Error searching communities", network.DB_ERROR, backupErr)
		}
	}

	aggregator.Close()
	return communitiesResults, nil
}
