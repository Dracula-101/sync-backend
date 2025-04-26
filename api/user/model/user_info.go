package model

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserInfo struct {
	Id                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserId            string             `bson:"userId" json:"userId"`
	Email             string             `bson:"email" json:"email"`
	FirstName         string             `bson:"firstName" json:"firstName"`
	LastName          string             `bson:"lastName" json:"lastName"`
	Bio               string             `bson:"bio" json:"bio"`
	VerifiedEmail     bool               `bson:"verifiedEmail" json:"verifiedEmail"`
	Avatar            UserAvatar         `bson:"avatar" json:"avatar"`
	Synergy           UserSynergy        `bson:"synergy" json:"synergy"`
	JoinedWavelengths []string           `bson:"joinedWavelengths" json:"joinedWavelengths"`
	ModeratorOf       []string           `bson:"moderatorOf" json:"moderatorOf"`
	Follows           int                `bson:"follows" json:"follows"`
	Followers         int                `bson:"followers" json:"followers"`
}

func (u *UserInfo) GetValue() *UserInfo {
	return u
}

func (u *UserInfo) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, err.Field()+" is required")
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}

type ProviderInfo struct {
	ProviderName string    `json:"name"`
	AddedAt      time.Time `json:"addedAt"`
}
