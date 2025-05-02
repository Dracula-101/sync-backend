package model

type UserSynergy struct {
	Posts    int `bson:"posts" json:"posts"`
	Comments int `bson:"comments" json:"comments"`
	Total    int `bson:"total" json:"total"`
}

func NewUserSynergy() UserSynergy {
	return UserSynergy{
		Posts:    0,
		Comments: 0,
		Total:    0,
	}
}
