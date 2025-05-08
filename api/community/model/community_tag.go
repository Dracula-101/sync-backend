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

type CommunityTagInfo struct {
	TagId       string `json:"id" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Icon        string `json:"icon" bson:"icon"`
}

func (tag *CommunityTag) ToCommunityTagInfo() CommunityTagInfo {
	return CommunityTagInfo{
		TagId:       tag.TagId,
		Name:        tag.Name,
		Description: tag.Description,
		Icon:        tag.Icon,
	}
}
