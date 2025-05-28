package model

import (
	user "sync-backend/api/user/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicGetCommunity struct {
	CommunityId string             `bson:"communityId" json:"id"`
	Slug        string             `bson:"slug" json:"slug"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	ShortDesc   string             `bson:"shortDesc" json:"shortDesc"`
	OwnerId     string             `bson:"ownerId" json:"ownerId"`
	IsJoined    bool               `bson:"isJoined" json:"isJoined"`
	IsPrivate   bool               `bson:"isPrivate" json:"isPrivate"`
	MemberCount int64              `bson:"memberCount" json:"memberCount"`
	PostCount   int64              `bson:"postCount" json:"postCount"`
	Media       CommunityMedia     `bson:"media" json:"media"`
	Rules       []CommunityRule    `bson:"rules" json:"rules"`
	Tags        []CommunityTagInfo `bson:"tags" json:"tags"`
	Moderators  []user.PublicUser  `bson:"moderators" json:"moderators"`
	Status      string             `bson:"status" json:"status"`
}

type PublicCommunity struct {
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Avatar      string             `json:"avatar"`
	Background  string             `json:"background"`
	CreatedAt   primitive.DateTime `json:"createdAt"`
	Status      CommunityStatus    `json:"status"`
}
