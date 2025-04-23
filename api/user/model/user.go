package model

import (
	"context"
	"sync-backend/arch/mongo"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const UserCollectionName = "users"

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name" validate:"required,max=200"`
	Email         string             `bson:"email" validate:"required,email"`
	Password      *string            `bson:"password" validate:"required,min=6,max=100"`
	ProfilePicURL string             `bson:"profilePicUrl,omitempty" validate:"omitempty,max=500"`
	Verified      bool               `bson:"verified" validate:"-"`
	Status        bool               `bson:"status" validate:"-"`
	Providers     []Provider         `bson:"providers,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" validate:"required"`
	UpdatedAt     time.Time          `bson:"updatedAt" validate:"required"`
}

func NewUser(email string, pwdHash string, name string, profilePicUrl string) (*User, error) {

	now := time.Now()
	u := User{
		Email:         email,
		Password:      &pwdHash,
		Name:          name,
		ProfilePicURL: profilePicUrl,
		Verified:      false,
		Status:        true,
		CreatedAt:     now,
		UpdatedAt:     now,
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
		UserId:     user.ID.Hex(),
		Name:       user.Name,
		Email:      user.Email,
		ProfilePic: user.ProfilePicURL,
		Providers:  providerInfo,
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
