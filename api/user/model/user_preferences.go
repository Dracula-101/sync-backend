package model

import "sync-backend/arch/common"

type UserPreferences struct {
	Language        common.LanguageDetail    `bson:"language" json:"language"`
	Theme           string                   `bson:"theme" json:"theme"`
	Timezone        common.TimeZoneDetail    `bson:"timezone" json:"timezone"`
	Location        string                   `bson:"location" json:"location"`
	Notifications   UserNotificationSettings `bson:"notifications" json:"notifications"`
	ContentSettings UserContentSettings      `bson:"contentSettings" json:"contentSettings"`
	PrivacySettings UserPrivacySettings      `bson:"privacySettings" json:"privacySettings"`
	BlockList       []string                 `bson:"blockList" json:"blockList"`
}

type UserNotificationSettings struct {
	Email bool `bson:"email" json:"email"`
	Push  bool `bson:"push" json:"push"`
}

type UserContentSettings struct {
	ShowSensitiveContent bool `bson:"showSensitiveContent" json:"showSensitiveContent"`
	ShowAdultContent     bool `bson:"showAdultContent" json:"showAdultContent"`
}

type UserPrivacySettings struct {
	IsProfileVisible           bool `bson:"isProfileVisible" json:"isProfileVisible"`
	IsEmailVisible             bool `bson:"isEmailVisible" json:"isEmailVisible"`
	IsJoinedWavelengthsVisible bool `bson:"isJoinedWavelengthsVisible" json:"isJoinedWavelengthsVisible"`
	FollowersVisible           bool `bson:"followersVisible" json:"followersVisible"`
}

func NewUserPreferences(
	language common.LanguageDetail,
	timezone common.TimeZoneDetail,
	theme string,
	location string,
) (up UserPreferences) {
	return UserPreferences{
		Language: language,
		Theme:    theme,
		Timezone: timezone,
		Location: location,
		Notifications: UserNotificationSettings{
			Email: true,
			Push:  false,
		},
		ContentSettings: UserContentSettings{
			ShowSensitiveContent: false,
			ShowAdultContent:     false,
		},
		PrivacySettings: UserPrivacySettings{
			IsProfileVisible:           true,
			IsEmailVisible:             true,
			IsJoinedWavelengthsVisible: true,
			FollowersVisible:           true,
		},
		BlockList: []string{},
	}
}
