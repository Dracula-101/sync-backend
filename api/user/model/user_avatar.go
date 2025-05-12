package model

type UserAvatar struct {
	ProfilePic Image `bson:"profilePic" json:"profilePic"`
	Background Image `bson:"background" json:"background"`
}

type Image struct {
	Id     string `bson:"id" json:"id"`
	Url    string `bson:"url" json:"url"`
	Width  int    `bson:"width" json:"width"`
	Height int    `bson:"height" json:"height"`
}

func NewUserAvatar(profileImage Image, backgroundImage Image) UserAvatar {
	return UserAvatar{
		ProfilePic: profileImage,
		Background: backgroundImage,
	}
}
