package model

const CommunityTagCollectionName = "community_tags"

type CommunityTag struct {
	Id          string `json:"id" bson:"_id,omitempty"`
	TagId       string `json:"tag_id" bson:"tag_id"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Icon        string `json:"icon" bson:"icon"`
	CreatedAt   string `json:"created_at,omitempty" bson:"created_at"`
}
