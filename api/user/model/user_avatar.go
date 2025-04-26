package model

import "sync-backend/utils"

type UserAvatar struct {
	ProfilePic Image `bson:"profilePic" json:"profilePic"`
	Background Image `bson:"background" json:"background"`
}

type Image struct {
	Url    string `bson:"url" json:"url"`
	Width  int    `bson:"width" json:"width"`
	Height int    `bson:"height" json:"height"`
}

func NewUserAvatar(profileImageUrl string, avatarImageUrl string) UserAvatar {
	profileImageWidth, profileImageHeight, err := utils.GetImageSize(profileImageUrl)
	if err != nil {
		profileImageHeight = 0
		profileImageWidth = 0
	}
	avatarImageWidth, avatarImageHeight, err := utils.GetImageSize(avatarImageUrl)
	if err != nil {
		avatarImageHeight = 0
		avatarImageWidth = 0
	}
	return UserAvatar{
		ProfilePic: Image{
			Url:    profileImageUrl,
			Width:  profileImageWidth,
			Height: profileImageHeight,
		},
		Background: Image{
			Url:    avatarImageUrl,
			Width:  avatarImageWidth,
			Height: avatarImageHeight,
		},
	}
}
