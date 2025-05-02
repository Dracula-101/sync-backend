package community

import (
	"errors"
	"sync-backend/api/community/model"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"go.mongodb.org/mongo-driver/bson"
)

const CommunityCollectionName = "community"
const CommunityTagCollectionName = "community_tags"

type CommunityService interface {
	CreateCommunity(name string, description string, tags []string, avatarUrl string, backgroundUrl string, userId string) (*model.Community, network.ApiError)
	GetCommunityById(id string) (*model.Community, network.ApiError)
}

type communityService struct {
	network.BaseService
	logger                   utils.AppLogger
	communityQueryBuilder    mongo.QueryBuilder[model.Community]
	communityTagQueryBuilder mongo.QueryBuilder[model.CommunityTag]
}

func NewCommunityService(db mongo.Database) CommunityService {
	return &communityService{
		BaseService:              network.NewBaseService(),
		logger:                   utils.NewServiceLogger("CommunityService"),
		communityQueryBuilder:    mongo.NewQueryBuilder[model.Community](db, CommunityCollectionName),
		communityTagQueryBuilder: mongo.NewQueryBuilder[model.CommunityTag](db, CommunityTagCollectionName),
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
	convertedTags := make([]model.CommunityTag, len(communityTags))
	for i, tag := range communityTags {
		convertedTags[i] = *tag
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
	if err != nil {
		s.logger.Error("Error fetching community: %v", err)
		return nil, network.NewInternalServerError("Error fetching community", network.DB_ERROR, err)
	}
	if community == nil {
		s.logger.Error("Community not found")
		return nil, network.NewNotFoundError("Community not found", errors.New("community not found"))
	}
	return community, nil
}
