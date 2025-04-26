package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/go-playground/validator/v10"
)

var (
	GoogleProviderName = "google"
)

type Provider struct {
	Id           primitive.ObjectID `bson:"_id,omitempty"`
	AuthIdToken  string             `bson:"idToken" validate:"required"`
	AuthProvider string             `bson:"providerName" validate:"required"`
	AddedAt      time.Time          `bson:"addedAt" validate:"required"`
}

func (authProvider *Provider) GetValue() *Provider {
	return authProvider
}

func (authProvider *Provider) Validate() error {
	validate := validator.New()
	return validate.Struct(authProvider)
}

func (authProvider *Provider) GetProviderInfo() *ProviderInfo {
	return &ProviderInfo{
		ProviderName: authProvider.AuthProvider,
		AddedAt:      authProvider.AddedAt,
	}
}
