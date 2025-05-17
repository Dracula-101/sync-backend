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

const CommentInteractionCollectionName = "comment_interactions"

type CommentInteraction struct {
	Id              primitive.ObjectID     `bson:"_id,omitempty" json:"-"`
	InteractionId   string                 `bson:"interactionId" json:"interactionId"`
	CommentId       string                 `bson:"commentId" json:"commentId" validate:"required"`
	UserId          string                 `bson:"userId" json:"userId" validate:"required"`
	InteractionType CommentInteractionType `bson:"interactionType" json:"interactionType" validate:"required,oneof=like dislike save"`
	CreatedAt       primitive.DateTime     `bson:"createdAt" json:"createdAt"`
	UpdatedAt       primitive.DateTime     `bson:"updatedAt" json:"updatedAt"`
	DeletedAt       *primitive.DateTime    `bson:"deletedAt,omitempty" json:"-"`
}

type CommentInteractionType string

const (
	CommentInteractionTypeLike    CommentInteractionType = "like"
	CommentInteractionTypeDislike CommentInteractionType = "dislike"
)

func NewCommentInteraction(userId string, commentId string, interactionType CommentInteractionType) *CommentInteraction {
	return &CommentInteraction{
		InteractionId:   uuid.NewString(),
		UserId:          userId,
		CommentId:       commentId,
		InteractionType: interactionType,
		CreatedAt:       primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:       primitive.NewDateTimeFromTime(time.Now()),
	}
}

func (c *CommentInteraction) GetValue() *CommentInteraction {
	return c
}

func (c *CommentInteraction) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func (c *CommentInteraction) GetCollectionName() string {
	return CommentInteractionCollectionName
}

func (*CommentInteraction) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "interactionId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_comment_interaction_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "commentId", Value: 1},
				{Key: "userId", Value: 1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_comment_interaction_unique"),
		},
		{
			Keys: bson.D{
				{Key: "commentId", Value: 1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_comment_interaction_comment"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_comment_interaction_user"),
		},
		{
			Keys: bson.D{
				{Key: "createdAt", Value: -1},
				{Key: "interactionType", Value: 1},
			},
			Options: options.Index().SetName("idx_comment_interaction_created"),
		},
		{
			// TTL index for soft-deleted records
			Keys: bson.D{
				{Key: "deletedAt", Value: 1},
			},
			Options: options.Index().SetName("idx_comment_interaction_deleted").SetExpireAfterSeconds(0),
		},
	}
	mongo.NewQueryBuilder[CommentInteraction](db, CommentInteractionCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
