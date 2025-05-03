package model

import (
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CommentCollectionName = "comments"

// Comment represents a user comment on a post or another comment
type Comment struct {
	Id            primitive.ObjectID  `bson:"_id,omitempty" json:"-"`
	CommentId     string              `bson:"commentId" json:"id"`
	PostId        string              `bson:"postId" json:"postId" validate:"required"`
	ParentId      string              `bson:"parentId,omitempty" json:"parentId"`
	AuthorId      string              `bson:"authorId" json:"authorId" validate:"required"`
	Content       string              `bson:"content" json:"content" validate:"required,max=10000"`
	Synergy       int                 `bson:"synergy" json:"synergy"`
	UpvoteCount   int                 `bson:"upvoteCount" json:"upvoteCount"`
	DownvoteCount int                 `bson:"downvoteCount" json:"downvoteCount"`
	ReplyCount    int                 `bson:"replyCount" json:"replyCount"`
	Level         int                 `bson:"level" json:"level"`
	IsEdited      bool                `bson:"isEdited" json:"isEdited"`
	IsDeleted     bool                `bson:"isDeleted" json:"isDeleted"`
	IsRemoved     bool                `bson:"isRemoved" json:"isRemoved"`
	IsStickied    bool                `bson:"isStickied" json:"isStickied"`
	Voters        map[string]VoteType `bson:"voters,omitempty" json:"voters,omitempty"`
	Media         []Media             `bson:"media,omitempty" json:"media,omitempty"`
	Metadata      Metadata            `bson:"metadata" json:"metadata"`
	CreatedAt     primitive.DateTime  `bson:"createdAt" json:"createdAt"`
	UpdatedAt     primitive.DateTime  `bson:"updatedAt" json:"updatedAt"`
	DeletedAt     *primitive.DateTime `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
	Path          string              `bson:"path" json:"path"` // Path used for efficient tree traversal (e.g., "root.123.456")
}

// CommentTree represents a hierarchical structure of comments with their replies
type CommentTree struct {
	Comment Comment       `bson:"comment" json:"comment"`
	Replies []CommentTree `bson:"replies,omitempty" json:"replies,omitempty"`
}

// NewComment creates a new comment with default values
func NewComment(postId string, authorId string, content string, parentId string, level int) *Comment {
	now := primitive.NewDateTimeFromTime(time.Now())

	path := postId
	if parentId != "" {
		// If this is a reply, construct the path by appending to the parent's path
		path = parentId + "." + utils.GenerateUUID()
	}

	return &Comment{
		Id:            primitive.NewObjectID(),
		CommentId:     utils.GenerateUUID(),
		PostId:        postId,
		ParentId:      parentId,
		AuthorId:      authorId,
		Content:       content,
		Synergy:       0,
		UpvoteCount:   0,
		DownvoteCount: 0,
		ReplyCount:    0,
		Level:         level,
		IsEdited:      false,
		IsDeleted:     false,
		IsRemoved:     false,
		IsStickied:    false,
		Voters:        make(map[string]VoteType),
		Path:          path,
		Metadata: Metadata{
			CreatedBy: authorId,
			Version:   1,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
