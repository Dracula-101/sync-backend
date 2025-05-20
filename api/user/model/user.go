package model

import (
	"context"
	"sync-backend/arch/common"
	"sync-backend/arch/mongo"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const UserCollectionName = "users"

type User struct {
	Id                primitive.ObjectID  `bson:"_id,omitempty" json:"-"`
	UserId            string              `bson:"userId" json:"id"`
	Username          string              `bson:"username" json:"username"`
	Email             string              `bson:"email" json:"email"`
	PasswordHash      string              `bson:"passwordHash" json:"-"`
	Bio               string              `bson:"bio" json:"bio"`
	VerifiedEmail     bool                `bson:"verifiedEmail" json:"verifiedEmail"`
	Status            UserStatus          `bson:"status" json:"status"`
	Avatar            UserAvatar          `bson:"avatar" json:"avatar"`
	Synergy           UserSynergy         `bson:"synergy" json:"synergy"`
	Providers         []Provider          `bson:"providers" json:"providers"`
	JoinedWavelengths []string            `bson:"joinedWavelengths" json:"joinedWavelengths"`
	ModeratorOf       []string            `bson:"moderatorOf" json:"moderatorOf"`
	Follows           []string            `bson:"follows" json:"follows"`
	Followers         []string            `bson:"followers" json:"followers"`
	Preferences       UserPreferences     `bson:"preferences" json:"preferences"`
	DeviceTokens      []DeviceToken       `bson:"deviceTokens" json:"-"`
	LoginHistory      []LoginHistory      `bson:"loginHistory" json:"-"`
	LastSeen          primitive.DateTime  `bson:"lastSeen" json:"lastSeen"`
	CreatedAt         primitive.DateTime  `bson:"createdAt" json:"createdAt"`
	UpdatedAt         primitive.DateTime  `bson:"updatedAt" json:"updatedAt"`
	DeletedAt         *primitive.DateTime `bson:"deletedAt,omitempty" json:"-"`
}

type UserStatus string

const (
	Active   UserStatus = "active"
	Inactive UserStatus = "inactive"
	Banned   UserStatus = "banned"
	Deleted  UserStatus = "deleted"
)

type NewUserArgs struct {
	UserName      string
	Email         string
	PasswordHash  string
	AvatarUrl     Image
	BackgroundUrl Image
	Language      common.Language
	TimeZone      common.TimeZone
	Theme         string
	Country       string
	DeviceToken   DeviceToken
}

func NewUser(
	newUserArgs NewUserArgs,
) (*User, error) {

	if newUserArgs.PasswordHash == "" {
		// Default password hash for empty password
		newUserArgs.PasswordHash = "$2a$10$Cv/Xb2ykZ9FLmWyB6vaPEueAzA51kkU2GDZj8C4hwgAH3gQhwIo.q"
	}
	now := time.Now()
	u := User{
		UserId:            uuid.New().String(),
		Username:          newUserArgs.UserName,
		Email:             newUserArgs.Email,
		PasswordHash:      newUserArgs.PasswordHash,
		VerifiedEmail:     false,
		Bio:               "",
		Status:            Active,
		Avatar:            NewUserAvatar(newUserArgs.AvatarUrl, newUserArgs.BackgroundUrl),
		Synergy:           NewUserSynergy(),
		Providers:         []Provider{},
		JoinedWavelengths: []string{},
		ModeratorOf:       []string{},
		Follows:           []string{},
		Followers:         []string{},
		Preferences: NewUserPreferences(
			UserPreferencesArgs{
				timezone: newUserArgs.TimeZone.ToDetail(),
				Language: newUserArgs.Language.ToDetail(),
				Theme:    newUserArgs.Theme,
				Location: newUserArgs.Country,
			},
		),
		DeviceTokens: []DeviceToken{
			newUserArgs.DeviceToken,
		},
		LoginHistory: []LoginHistory{},
		LastSeen:     primitive.NewDateTimeFromTime(now),
		CreatedAt:    primitive.NewDateTimeFromTime(now),
		UpdatedAt:    primitive.NewDateTimeFromTime(now),
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return &u, nil
}

func NewAuthProvider(authIdToken string, authProvider string, username string) (*Provider, error) {
	now := time.Now()
	p := Provider{
		AuthIdToken:  authIdToken,
		AuthProvider: authProvider,
		Username:     username,
		AddedAt:      now,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	p.Id = primitive.NewObjectID()
	return &p, nil
}

func (user *User) GetValue() *User {
	return user
}

func (user *User) GetCollectionName() string {
	return UserCollectionName
}

func (user *User) Validate() error {
	validate := validator.New()
	return validate.Struct(user)
}

func (user *User) GetUserInfo() *UserInfo {
	var providerInfo = make([]ProviderInfo, len(user.Providers))
	for i, provider := range user.Providers {
		providerInfo[i] = ProviderInfo{
			ProviderName: provider.AuthProvider,
			AddedAt:      provider.AddedAt,
			Username:     provider.Username,
		}
	}
	return &UserInfo{
		Username:          user.Username,
		UserId:            user.UserId,
		Email:             user.Email,
		Bio:               user.Bio,
		VerifiedEmail:     user.VerifiedEmail,
		Avatar:            user.Avatar,
		Synergy:           user.Synergy,
		JoinedWavelengths: user.JoinedWavelengths,
		ModeratorOf:       user.ModeratorOf,
		Follows:           len(user.Follows),
		Followers:         len(user.Followers),
	}
}

func (*User) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "email", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_user_email_unique"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_user_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "username", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_user_username_unique"),
		},
		{
			Keys: bson.D{
				{Key: "providers.authProvider", Value: 1},
				{Key: "providers.authIdToken", Value: 1},
			},
			Options: options.Index().SetName("idx_user_auth_providers"),
		},
		{
			Keys: bson.D{
				{Key: "lastSeen", Value: -1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_user_activity"),
		},
		// TTL index for deleted users - 30 days
		{
			Keys: bson.D{
				{Key: "deletedAt", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(30 * 24 * 60 * 60).SetName("ttl_user_deleted"),
		},
	}
	mongo.NewQueryBuilder[User](db, UserCollectionName).Query(context.Background()).CheckIndexes(indexes)

}
