package model

import (
	"context"
	"sync-backend/arch/mongo"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CommunityInteractionsCollectionName = "community_interactions"

// CommunityInteractionType defines the type of interaction with a community
type CommunityInteractionType string

const (
	CommunityInteractionTypeJoin  CommunityInteractionType = "join"
	CommunityInteractionTypeLeave CommunityInteractionType = "leave"
)

// CommunityInteraction represents a user interaction with a community
type CommunityInteraction struct {
	Id              primitive.ObjectID       `bson:"_id,omitempty" json:"-"`
	InteractionId   string                   `bson:"interactionId" json:"interactionId"`
	CommunityId     string                   `bson:"communityId" json:"communityId" validate:"required"`
	UserId          string                   `bson:"userId" json:"userId" validate:"required"`
	InteractionType CommunityInteractionType `bson:"interactionType" json:"interactionType" validate:"required,oneof=join leave"`
	DeviceInfo      DeviceInfo               `bson:"deviceInfo,omitempty" json:"deviceInfo,omitempty"`
	LocationInfo    *LocationInfo            `bson:"locationInfo,omitempty" json:"locationInfo,omitempty"`
	CreatedAt       primitive.DateTime       `bson:"createdAt" json:"createdAt"`
	UpdatedAt       primitive.DateTime       `bson:"updatedAt" json:"updatedAt"`
	DeletedAt       *primitive.DateTime      `bson:"deletedAt,omitempty" json:"-"`
}

// DeviceInfo represents information about the user's device
type DeviceInfo struct {
	DeviceId   string `bson:"deviceId,omitempty" json:"deviceId,omitempty"`
	DeviceType string `bson:"deviceType,omitempty" json:"deviceType,omitempty"`
	DeviceOS   string `bson:"deviceOS,omitempty" json:"deviceOS,omitempty"`
	AppVersion string `bson:"appVersion,omitempty" json:"appVersion,omitempty"`
}

// LocationInfo represents information about the user's location
type LocationInfo struct {
	Country   string  `bson:"country,omitempty" json:"country,omitempty"`
	City      string  `bson:"city,omitempty" json:"city,omitempty"`
	Latitude  float64 `bson:"latitude,omitempty" json:"latitude,omitempty"`
	Longitude float64 `bson:"longitude,omitempty" json:"longitude,omitempty"`
	IpAddress string  `bson:"ipAddress,omitempty" json:"ipAddress,omitempty"`
	Timezone  string  `bson:"timezone,omitempty" json:"timezone,omitempty"`
}

func NewCommunityInteraction(userId string, communityId string, interactionType CommunityInteractionType) *CommunityInteraction {
	return &CommunityInteraction{
		InteractionId:   uuid.NewString(),
		UserId:          userId,
		CommunityId:     communityId,
		InteractionType: interactionType,
		CreatedAt:       primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:       primitive.NewDateTimeFromTime(time.Now()),
	}
}

func (c *CommunityInteraction) GetValue() *CommunityInteraction {
	return c
}

func (c *CommunityInteraction) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func (c *CommunityInteraction) GetCollectionName() string {
	return CommunityInteractionsCollectionName
}

func (*CommunityInteraction) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "interactionId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_community_interaction_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
				{Key: "userId", Value: 1},
				{Key: "interactionType", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_community_interaction_lookup"),
		},
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_community_interaction_community"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_community_interaction_user"),
		},
		{
			Keys: bson.D{
				{Key: "createdAt", Value: -1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_community_interaction_created"),
		},
		{
			// TTL index for soft-deleted records
			Keys: bson.D{
				{Key: "deletedAt", Value: 1},
			},
			Options: options.Index().SetName("idx_community_interaction_deleted").SetExpireAfterSeconds(0),
		},
	}
	mongo.NewQueryBuilder[CommunityInteraction](db, CommunityInteractionsCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
