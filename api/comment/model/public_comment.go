package model

import (
	community "sync-backend/api/community/model"
	user "sync-backend/api/user/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicComment struct {
	Id             string                    `json:"id"`
	PostId         string                    `json:"post_id"`
	ParentId       string                    `json:"parent_id,omitempty"`
	Author         user.PublicUser           `json:"author"`
	Community      community.PublicCommunity `json:"community"`
	Content        string                    `json:"content"`
	Status         CommentStatus             `json:"status"`
	Synergy        int                       `json:"synergy"`
	ReplyCount     int                       `json:"reply_count"`
	ReactionCount  int                       `json:"reaction_count"`
	ReactionCounts map[ReactionType]int      `json:"reaction_counts"`
	Reactions      []Reaction                `json:"reactions"`
	Level          int                       `json:"level"`
	IsEdited       bool                      `json:"is_edited"`
	IsPinned       bool                      `json:"is_pinned"`
	IsLocked       bool                      `json:"is_locked"`
	IsArchived     bool                      `json:"is_archived"`
	IsDeleted      bool                      `json:"is_deleted"`
	CreatedAt      primitive.DateTime        `json:"created_at"`
}
