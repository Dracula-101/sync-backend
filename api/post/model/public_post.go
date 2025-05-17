package model

import (
	community "sync-backend/api/community/model"
	user "sync-backend/api/user/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicPost struct {
	Id           string                    `json:"id"`
	Title        string                    `json:"title"`
	Content      string                    `json:"content"`
	Author       user.PublicUser           `json:"author"`
	Community    community.PublicCommunity `json:"community"`
	Type         PostType                  `json:"type"`
	Status       PostStatus                `json:"status"`
	Media        []Media                   `json:"media,omitempty"`
	Tags         []string                  `json:"tags,omitempty"`
	Synergy      int                       `json:"synergy"`
	CommentCount int                       `json:"commentCount"`
	ViewCount    int                       `json:"viewCount"`
	ShareCount   int                       `json:"shareCount"`
	SaveCount    int                       `json:"saveCount"`
	Voters       map[string]VoteType       `json:"voters,omitempty"`
	IsNSFW       bool                      `json:"isNSFW"`
	IsSpoiler    bool                      `json:"isSpoiler"`
	IsStickied   bool                      `json:"isStickied"`
	IsLocked     bool                      `json:"isLocked"`
	IsArchived   bool                      `json:"isArchived"`
	CreatedAt    primitive.DateTime        `json:"createdAt"`
}

func (p *PublicPost) IsActive() bool {
	return p.Status == PostStatusActive
}
