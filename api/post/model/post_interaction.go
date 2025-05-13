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

const PostInteractionCollectionName = "post_interactions"

// Interaction represents a user interaction with a post
type PostInteraction struct {
	Id              primitive.ObjectID  `bson:"_id,omitempty" json:"-"`
	InteractionId   string              `bson:"interactionId" json:"interactionId"`
	PostId          string              `bson:"postId" json:"postId" validate:"required"`
	UserId          string              `bson:"userId" json:"userId" validate:"required"`
	InteractionType InteractionType     `bson:"interactionType" json:"interactionType" validate:"required,oneof=like dislike save share comment"`
	CreatedAt       primitive.DateTime  `bson:"createdAt" json:"createdAt"`
	UpdatedAt       primitive.DateTime  `bson:"updatedAt" json:"updatedAt"`
	DeletedAt       *primitive.DateTime `bson:"deletedAt,omitempty" json:"-"`
}

// InteractionType defines the type of interaction
type InteractionType string

const (
	InteractionTypeLike    InteractionType = "like"
	InteractionTypeDislike InteractionType = "dislike"
	InteractionTypeSave    InteractionType = "save"
)

func NewPostInteraction(userId string, postId string, interactionType InteractionType) *PostInteraction {
	return &PostInteraction{
		InteractionId:   uuid.NewString(),
		UserId:          userId,
		PostId:          postId,
		InteractionType: interactionType,
		CreatedAt:       primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:       primitive.NewDateTimeFromTime(time.Now()),
	}
}

func (p *PostInteraction) GetValue() *PostInteraction {
	return p
}

func (p *PostInteraction) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

func (p *PostInteraction) GetCollectionName() string {
	return PostInteractionCollectionName
}

func (*PostInteraction) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "interactionId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_interaction_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "postId", Value: 1},
				{Key: "userId", Value: 1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_post_interaction_unique"),
		},
		{
			Keys: bson.D{
				{Key: "postId", Value: 1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_post_interaction_post"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_post_interaction_user"),
		},
		{
			Keys: bson.D{
				{Key: "createdAt", Value: -1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_post_interaction_created"),
		},
		{
			//ttl
			Keys: bson.D{
				{Key: "deletedAt", Value: 1},
			},
			Options: options.Index().SetName("idx_post_interaction_deleted").SetExpireAfterSeconds(0),
		},
	}
	mongo.NewQueryBuilder[PostInteraction](db, PostInteractionCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
