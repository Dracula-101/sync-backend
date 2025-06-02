package model

import (
	"sync-backend/api/community/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeedPost struct {
	ID           string                `json:"id" bson:"postId"`
	Title        string                `json:"title" bson:"title"`
	Content      string                `json:"content" bson:"content"`
	AuthorId     string                `json:"authorId" bson:"authorId"`
	Community    model.PublicCommunity `json:"community" bson:"community"`
	Type         PostType              `json:"type" bson:"type"`
	Status       PostStatus            `json:"status" bson:"status"`
	Tags         []string              `json:"tags" bson:"tags"`
	Synergy      int                   `json:"synergy" bson:"synergy"`
	IsLiked      bool                  `json:"isLiked" bson:"isLiked"`
	IsDisliked   bool                  `json:"isDisliked" bson:"isDisliked"`
	CommentCount int                   `json:"commentCount" bson:"commentCount"`
	ViewCount    int                   `json:"viewCount" bson:"viewCount"`
	ShareCount   int                   `json:"shareCount" bson:"shareCount"`
	SaveCount    int                   `json:"saveCount" bson:"saveCount"`
	IsNSFW       bool                  `json:"isNSFW" bson:"isNSFW"`
	IsSpoiler    bool                  `json:"isSpoiler" bson:"isSpoiler"`
	IsStickied   bool                  `json:"isStickied" bson:"isStickied"`
	IsLocked     bool                  `json:"isLocked" bson:"isLocked"`
	CreatedAt    primitive.DateTime    `json:"createdAt" bson:"createdAt"`
	UpdatedAt    primitive.DateTime    `json:"updatedAt" bson:"updatedAt"`
}
