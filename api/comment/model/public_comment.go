package model

import (
	community "sync-backend/api/community/model"
	user "sync-backend/api/user/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicComment struct {
	Id               string                    `json:"id"`
	PostId           string                    `json:"postId"`
	ParentId         string                    `json:"parentId,omitempty"`
	Author           user.PublicUser           `json:"author"`
	Community        community.PublicCommunity `json:"community"`
	Content          string                    `json:"content"`
	FormattedContent string                    `json:"formattedContent,omitempty"`
	Status           CommentStatus             `json:"status"`
	Synergy          int                       `json:"synergy"` // Overall score
	ReplyCount       int                       `json:"replyCount"`
	ReactionCounts   map[ReactionType]int      `json:"reactionCounts,omitempty"` // Count by reaction type
	Level            int                       `json:"level"`                    // Nesting level (0 for top-level)
	IsEdited         bool                      `json:"isEdited"`
	IsPinned         bool                      `json:"isPinned"`   // Pinned by author
	IsStickied       bool                      `json:"isStickied"` // Stickied by moderator
	IsLocked         bool                      `json:"isLocked"`   // Can't be replied to
	IsDeleted        bool                      `json:"isDeleted"`  // Soft delete by user
	IsRemoved        bool                      `json:"isRemoved"`  // Removed by moderator
	HasMedia         bool                      `json:"hasMedia"`
	Mentions         []string                  `json:"mentions,omitempty"`
	Path             string                    `json:"path"`
	CreatedAt        primitive.DateTime        `json:"createdAt"`
}
