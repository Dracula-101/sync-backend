package model

import (
	"context"
	"strings"
	"sync-backend/arch/mongo"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CommentCollectionName = "comments"

// CommentStatus defines the current status of a comment
type CommentStatus string

const (
	CommentStatusActive   CommentStatus = "active"
	CommentStatusPending  CommentStatus = "pending" // For moderation queue
	CommentStatusRemoved  CommentStatus = "removed" // By moderator
	CommentStatusDeleted  CommentStatus = "deleted" // By user
	CommentStatusArchived CommentStatus = "archived"
	CommentStatusHidden   CommentStatus = "hidden"  // Auto-hidden due to low score
	CommentStatusFlagged  CommentStatus = "flagged" // Flagged for review
)

// VoteType represents the type of vote a user has cast on a comment
type VoteType int

const (
	Downvote VoteType = -1
	NoVote   VoteType = 0
	Upvote   VoteType = 1
)

// Metadata represents common metadata fields for the comment
type Metadata struct {
	CreatedBy  string         `bson:"createdBy" json:"createdBy"`
	UpdatedBy  string         `bson:"updatedBy,omitempty" json:"updatedBy,omitempty"`
	DeletedBy  string         `bson:"deletedBy,omitempty" json:"-"`
	IPAddress  string         `bson:"ipAddress,omitempty" json:"-"`
	UserAgent  string         `bson:"userAgent,omitempty" json:"-"`
	Version    int            `bson:"version" json:"version"`
	CustomData map[string]any `bson:"customData,omitempty" json:"customData,omitempty"`
	ModNote    string         `bson:"modNote,omitempty" json:"-"` // Moderator's note
}

// ReactionType defines the type of reaction to a comment
type ReactionType string

const (
	ReactionTypeLike    ReactionType = "like"
	ReactionTypeLove    ReactionType = "love"
	ReactionTypeLaugh   ReactionType = "laugh"
	ReactionTypeSad     ReactionType = "sad"
	ReactionTypeAngry   ReactionType = "angry"
	ReactionTypeWow     ReactionType = "wow"
	ReactionTypeSupport ReactionType = "support"
)

// Reaction represents a user's reaction to a comment
type Reaction struct {
	UserId    string             `bson:"userId" json:"userId"`
	Type      ReactionType       `bson:"type" json:"type"`
	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
}

// DeviceInfo represents information about the device used to create the comment
type DeviceInfo struct {
	DeviceId   string `bson:"deviceId,omitempty" json:"-"`
	DeviceType string `bson:"deviceType,omitempty" json:"-"`
	DeviceOS   string `bson:"deviceOS,omitempty" json:"-"`
	AppVersion string `bson:"appVersion,omitempty" json:"-"`
}

// LocationInfo represents geographical information about where the comment was created
type LocationInfo struct {
	Country   string  `bson:"country,omitempty" json:"-"`
	City      string  `bson:"city,omitempty" json:"-"`
	Latitude  float64 `bson:"latitude,omitempty" json:"-"`
	Longitude float64 `bson:"longitude,omitempty" json:"-"`
	IpAddress string  `bson:"ipAddress,omitempty" json:"-"`
	Timezone  string  `bson:"timezone,omitempty" json:"-"`
}

// ModerationInfo represents moderation-related information for a comment
type ModerationInfo struct {
	ModeratorId     string              `bson:"moderatorId,omitempty" json:"-"`
	ReportCount     int                 `bson:"reportCount" json:"reportCount"`
	IsAutoModerated bool                `bson:"isAutoModerated" json:"-"`
	ModeratedAt     *primitive.DateTime `bson:"moderatedAt,omitempty" json:"-"`
	ModReason       string              `bson:"modReason,omitempty" json:"-"`
	LastReportedAt  *primitive.DateTime `bson:"lastReportedAt,omitempty" json:"-"`
	ToxicityScore   float64             `bson:"toxicityScore,omitempty" json:"-"` // For AI moderation
}

// Comment represents a user comment in the system
type Comment struct {
	Id               primitive.ObjectID   `bson:"_id,omitempty" json:"-"`
	CommentId        string               `bson:"commentId" json:"id"`
	PostId           string               `bson:"postId" json:"postId" validate:"required"` // ID of the post this comment belongs to
	ParentId         string               `bson:"parentId,omitempty" json:"parentId"`       // For nested comments/replies
	AuthorId         string               `bson:"authorId" json:"authorId" validate:"required"`
	CommunityId      string               `bson:"communityId" json:"communityId" validate:"required"` // Important for community-specific moderation
	Content          string               `bson:"content" json:"content" validate:"required,max=10000"`
	FormattedContent string               `bson:"formattedContent,omitempty" json:"formattedContent,omitempty"` // HTML or other formatted version
	RawContent       string               `bson:"rawContent,omitempty" json:"-"`                                // For storing original markdown or other format
	Status           CommentStatus        `bson:"status" json:"status"`
	Synergy          int                  `bson:"synergy" json:"synergy"` // Overall score
	ReplyCount       int                  `bson:"replyCount" json:"replyCount"`
	ReactionCounts   map[ReactionType]int `bson:"reactionCounts,omitempty" json:"reactionCounts,omitempty"` // Count by reaction type
	Reactions        []Reaction           `bson:"reactions,omitempty" json:"-"`
	Voters           map[string]VoteType  `bson:"voters,omitempty" json:"-"` // Map of userId to vote type
	Level            int                  `bson:"level" json:"level"`        // Nesting level (0 for top-level)
	IsEdited         bool                 `bson:"isEdited" json:"isEdited"`
	IsPinned         bool                 `bson:"isPinned" json:"isPinned"`     // Pinned by author
	IsStickied       bool                 `bson:"isStickied" json:"isStickied"` // Stickied by moderator
	IsLocked         bool                 `bson:"isLocked" json:"isLocked"`     // Can't be replied to
	IsDeleted        bool                 `bson:"isDeleted" json:"isDeleted"`   // Soft delete by user
	IsRemoved        bool                 `bson:"isRemoved" json:"isRemoved"`   // Removed by moderator
	HasMedia         bool                 `bson:"hasMedia" json:"hasMedia"`
	Mentions         []string             `bson:"mentions,omitempty" json:"mentions,omitempty"` // User IDs mentioned in the comment
	Metadata         Metadata             `bson:"metadata" json:"-"`
	DeviceInfo       DeviceInfo           `bson:"deviceInfo,omitempty" json:"-"`
	LocationInfo     *LocationInfo        `bson:"locationInfo,omitempty" json:"-"`
	ModerationInfo   ModerationInfo       `bson:"moderationInfo,omitempty" json:"-"`
	Path             string               `bson:"path" json:"path"` // Path for efficient tree traversal (e.g., "root.123.456")
	CreatedAt        primitive.DateTime   `bson:"createdAt" json:"createdAt"`
	UpdatedAt        primitive.DateTime   `bson:"updatedAt" json:"updatedAt"`
	DeletedAt        *primitive.DateTime  `bson:"deletedAt,omitempty" json:"-"`
	EditHistory      []CommentEdit        `bson:"editHistory,omitempty" json:"-"` // Track edits for moderation purposes
	Flags            map[string]bool      `bson:"flags,omitempty" json:"-"`       // For feature flags or special attributes
}

// CommentEdit represents a record of an edit made to a comment
type CommentEdit struct {
	EditorId  string             `bson:"editorId" json:"-"`
	Content   string             `bson:"content" json:"-"`
	Reason    string             `bson:"reason,omitempty" json:"-"`
	EditedAt  primitive.DateTime `bson:"editedAt" json:"-"`
	IPAddress string             `bson:"ipAddress,omitempty" json:"-"`
}

// CommentTree represents a hierarchical structure of comments with their replies
type CommentTree struct {
	Comment Comment       `bson:"comment" json:"comment"`
	Replies []CommentTree `bson:"replies,omitempty" json:"replies,omitempty"`
}

// NewComment creates a new comment with default values
func NewComment(postId string, authorId string, communityId string, content string, parentId string) *Comment {
	now := primitive.NewDateTimeFromTime(time.Now())
	commentId := uuid.New().String()

	// Calculate path for efficient tree traversal
	path := postId
	if parentId != "" {
		// If this is a reply, construct the path by appending to the parent's path
		path = parentId + "." + commentId
	}

	// generate the level based on the number of dots in the path
	level := len(strings.Split(path, ".")) - 1

	return &Comment{
		Id:             primitive.NewObjectID(),
		CommentId:      commentId,
		PostId:         postId,
		CommunityId:    communityId,
		ParentId:       parentId,
		AuthorId:       authorId,
		Content:        content,
		Status:         CommentStatusActive,
		Synergy:        0,
		ReplyCount:     0,
		Level:          level,
		IsEdited:       false,
		IsPinned:       false,
		IsStickied:     false,
		IsLocked:       false,
		IsDeleted:      false,
		IsRemoved:      false,
		Voters:         make(map[string]VoteType),
		ReactionCounts: make(map[ReactionType]int),
		Path:           path,
		Metadata: Metadata{
			CreatedBy: authorId,
			Version:   1,
		},
		ModerationInfo: ModerationInfo{
			ReportCount: 0,
		},
		Flags:     make(map[string]bool),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Comment) AddDeviceInfo(deviceId string, deviceType string, deviceOS string, appVersion string) {
	c.DeviceInfo.DeviceId = deviceId
	c.DeviceInfo.DeviceType = deviceType
	c.DeviceInfo.DeviceOS = deviceOS
	c.DeviceInfo.AppVersion = appVersion
}

func (c *Comment) AddLocationInfo(country string, city string, latitude float64, longitude float64, ip string, timezone string) {
	c.LocationInfo = &LocationInfo{
		Country:   country,
		City:      city,
		Latitude:  latitude,
		Longitude: longitude,
		IpAddress: ip,
		Timezone:  timezone,
	}
}

func (c *Comment) GetValue() *Comment {
	return c
}

func (c *Comment) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func (c *Comment) GetCollectionName() string {
	return CommentCollectionName
}

// EnsureIndexes creates all necessary indexes for the Comment model
func (*Comment) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "commentId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_comment_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "postId", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_comment_post_timeline"),
		},
		{
			Keys: bson.D{
				{Key: "parentId", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_comment_parent_timeline"),
		},
		{
			Keys: bson.D{
				{Key: "authorId", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_comment_author"),
		},
		{
			Keys: bson.D{
				{Key: "path", Value: 1},
			},
			Options: options.Index().SetName("idx_comment_path"),
		},
		{
			Keys: bson.D{
				{Key: "synergy", Value: -1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_comment_synergy"),
		},
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_comment_community"),
		},
		// TTL index for deleted comments - 7 days
		{
			Keys: bson.D{
				{Key: "deletedAt", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(7 * 24 * 60 * 60).SetName("ttl_comment_deleted"),
		},
	}

	mongo.NewQueryBuilder[Comment](db, CommentCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
