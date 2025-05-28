package model

type CommunitySearchResult struct {
	CommunityId    string   `bson:"communityId" json:"id"`
	Slug           string   `bson:"slug" json:"slug"`
	Name           string   `bson:"name" json:"name"`
	Description    string   `bson:"description" json:"description"`
	ShortDesc      string   `bson:"shortDesc" json:"shortDesc"`
	OwnerId        string   `bson:"ownerId" json:"ownerId"`
	IsPrivate      bool     `bson:"isPrivate" json:"isPrivate"`
	Members        []string `bson:"members" json:"members"`
	MemberCount    int64    `bson:"memberCount" json:"memberCount"`
	PostCount      int64    `bson:"postCount" json:"postCount"`
	Status         string   `bson:"status" json:"status"`
	Score          float64  `bson:"score,omitempty" json:"score,omitempty"`
	RelevanceScore float64  `bson:"relevanceScore,omitempty" json:"relevanceScore,omitempty"`
	Matched        []string `bson:"matched,omitempty" json:"matched,omitempty"`
}
