package model

import (
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const PostCollectionName = "posts"

// Post represents a user post in the system, similar to a Reddit post
type Post struct {
	Id             primitive.ObjectID  `bson:"_id,omitempty" json:"-"`
	PostId         string              `bson:"postId" json:"id"`
	Title          string              `bson:"title" json:"title" validate:"required,min=1,max=300"`
	Content        string              `bson:"content" json:"content"`
	AuthorId       string              `bson:"authorId" json:"authorId" validate:"required"`
	CommunityId    string              `bson:"communityId" json:"communityId" validate:"required"`
	Type           PostType            `bson:"type" json:"type" validate:"required,oneof=text image video link poll gallery"`
	Status         PostStatus          `bson:"status" json:"status"`
	Media          []Media             `bson:"media,omitempty" json:"media,omitempty"`
	Tags           []string            `bson:"tags,omitempty" json:"tags,omitempty"`
	Synergy        int                 `bson:"synergy" json:"synergy"`
	CommentCount   int                 `bson:"commentCount" json:"commentCount"`
	ViewCount      int                 `bson:"viewCount" json:"viewCount"`
	ShareCount     int                 `bson:"shareCount" json:"shareCount"`
	SaveCount      int                 `bson:"saveCount" json:"saveCount"`
	Voters         map[string]VoteType `bson:"voters,omitempty" json:"voters,omitempty"`
	IsNSFW         bool                `bson:"isNSFW" json:"isNSFW"`
	IsSpoiler      bool                `bson:"isSpoiler" json:"isSpoiler"`
	IsStickied     bool                `bson:"isStickied" json:"isStickied"`
	IsLocked       bool                `bson:"isLocked" json:"isLocked"`
	IsArchived     bool                `bson:"isArchived" json:"isArchived"`
	Metadata       Metadata            `bson:"metadata" json:"metadata"`
	CreatedAt      primitive.DateTime  `bson:"createdAt" json:"createdAt"`
	UpdatedAt      primitive.DateTime  `bson:"updatedAt" json:"updatedAt"`
	DeletedAt      *primitive.DateTime `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
	LastActivityAt primitive.DateTime  `bson:"lastActivityAt" json:"lastActivityAt"`
}

// Metadata represents common metadata fields used across models
type Metadata struct {
	CreatedBy  string         `bson:"createdBy" json:"createdBy"`
	UpdatedBy  string         `bson:"updatedBy,omitempty" json:"updatedBy,omitempty"`
	DeletedBy  string         `bson:"deletedBy,omitempty" json:"-"`
	IPAddress  string         `bson:"ipAddress,omitempty" json:"-"`
	UserAgent  string         `bson:"userAgent,omitempty" json:"-"`
	Version    int            `bson:"version" json:"version"`
	CustomData map[string]any `bson:"customData,omitempty" json:"customData,omitempty"`
}

// PostType defines the type of post
type PostType string

const (
	TextPost  PostType = "text"
	ImagePost PostType = "image"
	VideoPost PostType = "video"
	LinkPost  PostType = "link"
)

// PostStatus defines the current status of a post
type PostStatus string

const (
	PostStatusActive   PostStatus = "active"
	PostStatusPending  PostStatus = "pending"
	PostStatusRemoved  PostStatus = "removed"
	PostStatusDeleted  PostStatus = "deleted"
	PostStatusArchived PostStatus = "archived"
)

// VoteType represents the type of vote a user has cast on a post
type VoteType int

const (
	Downvote VoteType = -1
	NoVote   VoteType = 0
	Upvote   VoteType = 1
)

// Media represents media files attached to a post
type Media struct {
	Id           string             `bson:"id" json:"id"`
	Type         MediaType          `bson:"type" json:"type"`
	Url          string             `bson:"url" json:"url"`
	ThumbnailUrl string             `bson:"thumbnailUrl,omitempty" json:"thumbnailUrl,omitempty"`
	Width        int                `bson:"width,omitempty" json:"width,omitempty"`
	Height       int                `bson:"height,omitempty" json:"height,omitempty"`
	FileSize     int64              `bson:"fileSize,omitempty" json:"fileSize,omitempty"`
	MimeType     string             `bson:"mimeType,omitempty" json:"mimeType,omitempty"`
	Duration     int                `bson:"duration,omitempty" json:"duration,omitempty"`
	AltText      string             `bson:"altText,omitempty" json:"altText,omitempty"`
	CreatedAt    primitive.DateTime `bson:"createdAt" json:"createdAt"`
}

// MediaType defines the type of media
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeGIF   MediaType = "gif"
	MediaTypeAudio MediaType = "audio"
	MediaTypeFile  MediaType = "file"
)

// NewPost creates a new post with default values
func NewPost(authorId string, communityId string, title string, content string, tags []string, media []string, postType PostType, isNSFW bool, isSpoiler bool) *Post {
	now := primitive.NewDateTimeFromTime(time.Now())

	return &Post{
		Id:           primitive.NewObjectID(),
		PostId:       utils.GenerateUUID(),
		Title:        title,
		Content:      content,
		AuthorId:     authorId,
		CommunityId:  communityId,
		Type:         postType,
		Status:       PostStatusActive,
		Tags:         tags,
		Synergy:      0,
		CommentCount: 0,
		ViewCount:    0,
		ShareCount:   0,
		SaveCount:    0,
		Voters:       make(map[string]VoteType),
		IsNSFW:       false,
		IsSpoiler:    false,
		IsStickied:   false,
		IsLocked:     false,
		IsArchived:   false,
		Metadata: Metadata{
			CreatedBy: authorId,
			Version:   1,
		},
		CreatedAt:      now,
		UpdatedAt:      now,
		LastActivityAt: now,
	}
}
