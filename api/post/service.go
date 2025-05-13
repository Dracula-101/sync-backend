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
	GetPost(postId string) (*model.Post, error)
	EditPost(userId string, postId string, title *string, content *string, postType model.PostType, isNSFW *bool, isSpoiler *bool) (*string, error)
	GetPostsByUserId(userId string, page int, limit int) (posts []*model.Post, numOfPosts int, err error)
	GetPostsByCommunityId(communityId string, page int, limit int) (posts []*model.Post, numOfPosts int, err error)
}

type postService struct {
	network.BaseService
	mediaService     media.MediaService
	userService      user.UserService
	logger           utils.AppLogger
	communityService community.CommunityService
	postQueryBuilder mongo.QueryBuilder[model.Post]
}

func NewPostService(db mongo.Database, userService user.UserService, communityService community.CommunityService, mediaService media.MediaService) PostService {
	return &postService{
		BaseService:      network.NewBaseService(),
		logger:           utils.NewServiceLogger("PostService"),
		mediaService:     mediaService,
		userService:      userService,
		communityService: communityService,
		postQueryBuilder: mongo.NewQueryBuilder[model.Post](db, model.PostCollectionName),
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

func (s *postService) GetPost(postId string) (*model.Post, error) {
	s.logger.Info("Getting post with ID: %s", postId)
	filter := bson.M{"postId": postId}
	post, err := s.postQueryBuilder.SingleQuery().FindOne(filter, nil)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		s.logger.Error("Failed to get post: %v", err)
		return nil, network.NewInternalServerError("Failed to get post", network.DB_ERROR, err)
	}
	if post == nil {
		s.logger.Error("Post not found")
		return nil, network.NewNotFoundError("Post not found", fmt.Errorf("post with ID %s not found", postId))
	}
	s.logger.Info("Post retrieved successfully with ID: %s", post.PostId)
	return post, nil
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

func (s *postService) EditPost(userId string, postId string, title *string, content *string, postType model.PostType, isNSFW *bool, isSpoiler *bool) (newPostId *string, err error) {
	s.logger.Info("Editing post with ID: %s", postId)
	post, err := s.GetPost(postId)
	if err != nil {
		return nil, err
	}
	if !post.IsActive() {
		s.logger.Error("Cannot edit inactive post with ID: %s", postId)
		return nil, network.NewForbiddenError("Cannot edit inactive post", fmt.Errorf("post with ID %s is not active", postId))
	}
	if post.AuthorId != userId {
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
