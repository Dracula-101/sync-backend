package post

import (
	"sync-backend/api/community"
	"sync-backend/api/post/model"
	"sync-backend/api/user"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"
)

type PostService interface {
	CreatePost(title string, content string, tags []string, media []string, userId string, communityId string, postType model.PostType, isNSFW bool, isSpoiler bool) (*model.Post, error)
}

type postService struct {
	network.BaseService
	logger           utils.AppLogger
	userService      user.UserService
	communityService community.CommunityService
	postQueryBuilder mongo.QueryBuilder[model.Post]
}

func NewPostService(db mongo.Database, userService user.UserService, communityService community.CommunityService) PostService {
	return &postService{
		BaseService:      network.NewBaseService(),
		logger:           utils.NewServiceLogger("PostService"),
		userService:      userService,
		communityService: communityService,
		postQueryBuilder: mongo.NewQueryBuilder[model.Post](db, model.PostCollectionName),
	}
}

func (s *postService) CreatePost(
	title string, content string, tags []string, media []string, userId string, communityId string, postType model.PostType, isNSFW bool, isSpoiler bool,
) (*model.Post, error) {
	s.logger.Info("Creating post with title: %s", title)
	post := model.NewPost(userId, communityId, title, content, tags, media, postType, isNSFW, isSpoiler)

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
