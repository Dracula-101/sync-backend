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

// Represents a unique collection name for community bans
const CommunityBansCollectionName = "community_bans"

// CommunityBan represents a user ban record in a community
type CommunityBan struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"-"`
	BanId       string              `bson:"banId" json:"id"`
	CommunityId string              `bson:"communityId" json:"communityId" validate:"required"`
	UserId      string              `bson:"userId" json:"userId" validate:"required"`
	ModeratorId string              `bson:"moderatorId" json:"moderatorId" validate:"required"`
	Reason      string              `bson:"reason" json:"reason"`
	Duration    *int                `bson:"duration,omitempty" json:"duration,omitempty"`
	ExpiresAt   *primitive.DateTime `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
	IsActive    bool                `bson:"isActive" json:"isActive"`
	CreatedAt   primitive.DateTime  `bson:"createdAt" json:"createdAt"`
	UpdatedAt   primitive.DateTime  `bson:"updatedAt" json:"updatedAt"`
}

// NewCommunityBan creates a new community ban
func NewCommunityBan(communityId, userId, moderatorId, reason string, duration *int) *CommunityBan {
	now := primitive.NewDateTimeFromTime(time.Now())
	banId := uuid.New().String()

	var expiresAt *primitive.DateTime
	if duration != nil {
		expTime := primitive.NewDateTimeFromTime(now.Time().Add(time.Duration(*duration) * time.Hour))
		expiresAt = &expTime
	}

	return &CommunityBan{
		BanId:       banId,
		CommunityId: communityId,
		UserId:      userId,
		ModeratorId: moderatorId,
		Reason:      reason,
		Duration:    duration,
		ExpiresAt:   expiresAt,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

type BanInfo struct {
	Reason      string     `json:"reason"`
	IsPermanent bool       `json:"isPermanent"`
	Duration    *int       `json:"duration,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

func (b *CommunityBan) GetValue() *CommunityBan {
	return b
}

func (b *CommunityBan) Validate() error {
	validate := validator.New()
	return validate.Struct(b)
}

func (b *CommunityBan) GetCollectionName() string {
	return CommunityBansCollectionName
}

func (b *CommunityBan) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
				{Key: "userId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "expiresAt", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}
	mongo.NewQueryBuilder[CommunityBan](db, CommunityBansCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
