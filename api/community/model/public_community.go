package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type PublicCommunity struct {
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Avatar      string             `json:"avatar"`
	Background  string             `json:"background"`
	CreatedAt   primitive.DateTime `json:"createdAt"`
	Status      CommunityStatus    `json:"status"`
}
