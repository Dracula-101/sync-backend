package model

import (
	"context"
	"sync-backend/arch/common"
	"sync-backend/arch/mongo"
	"sync-backend/utils"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const UserCollectionName = "users"

type User struct {
	Id                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserId            string             `bson:"userId" json:"userId"`
	Email             string             `bson:"email" json:"email"`
	PasswordHash      string             `bson:"passwordHash" json:"passwordHash"`
	FirstName         string             `bson:"firstName" json:"firstName"`
	LastName          string             `bson:"lastName" json:"lastName"`
	Bio               string             `bson:"bio" json:"bio"`
	VerifiedEmail     bool               `bson:"verifiedEmail" json:"verifiedEmail"`
	Status            UserStatus         `bson:"status" json:"status"`
	Avatar            UserAvatar         `bson:"avatar" json:"avatar"`
	Synergy           UserSynergy        `bson:"synergy" json:"synergy"`
	Providers         []Provider         `bson:"providers" json:"providers"`
	JoinedWavelengths []string           `bson:"joinedWavelengths" json:"joinedWavelengths"`
	ModeratorOf       []string           `bson:"moderatorOf" json:"moderatorOf"`
	Follows           []string           `bson:"follows" json:"follows"`
	Followers         []string           `bson:"followers" json:"followers"`
	Preferences       UserPreferences    `bson:"preferences" json:"preferences"`
	DeviceTokens      []DeviceToken      `bson:"deviceTokens" json:"deviceTokens"`
	LoginHistory      []LoginHistory     `bson:"loginHistory" json:"loginHistory"`
	LastSeen          primitive.DateTime `bson:"lastSeen" json:"lastSeen"`
	CreatedAt         primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt         primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
	DeletedAt         primitive.DateTime `bson:"deletedAt" json:"deletedAt"`
}

type UserStatus string

const (
	Active      UserStatus = "active"
	Inactive    UserStatus = "inactive"
	UnAvailable UserStatus = "unavailable"
)

func NewUser(
	email string,
	passwordHash string,
	firstName string,
	lastName string,
	bio string,
	avatarUrl string,
	backgroundUrl string,
	language common.Language,
	timeZone common.TimeZone,
	deviceToken DeviceToken,
) (*User, error) {

	now := time.Now()
	u := User{
		Id:                primitive.NewObjectID(),
		UserId:            utils.GenerateUUID(),
		Email:             email,
		PasswordHash:      passwordHash,
		FirstName:         firstName,
		LastName:          lastName,
		VerifiedEmail:     false,
		Bio:               bio,
		Status:            Active,
		Avatar:            NewUserAvatar(avatarUrl, backgroundUrl),
		Synergy:           NewUserSynergy(),
		Providers:         []Provider{},
		JoinedWavelengths: []string{},
		ModeratorOf:       []string{},
		Follows:           []string{},
		Followers:         []string{},
		Preferences:       NewUserPreferences(language.ToDetail(), timeZone.ToDetail(), "dark", "India"),
		DeviceTokens:      []DeviceToken{deviceToken},
		LoginHistory:      []LoginHistory{},
		LastSeen:          primitive.NewDateTimeFromTime(now),
		CreatedAt:         primitive.NewDateTimeFromTime(now),
		UpdatedAt:         primitive.NewDateTimeFromTime(now),
		DeletedAt:         primitive.NewDateTimeFromTime(now),
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return &u, nil
}

func NewAuthProvider(authIdToken string, authProvider string) (*Provider, error) {
	now := time.Now()
	p := Provider{
		AuthIdToken:  authIdToken,
		AuthProvider: authProvider,
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
		}
	}
	return &UserInfo{
		Id:                user.Id,
		UserId:            user.UserId,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
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
				{Key: "_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "email", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "email", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	mongo.NewQueryBuilder[User](db, UserCollectionName).Query(context.Background()).CreateIndexes(indexes)
}
