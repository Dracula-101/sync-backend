package model

import (
	"context"
	"sync-backend/arch/mongo"
	"sync-backend/utils"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const SessionCollectionName = "sessions"

// Session represents a user session
type Session struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SessionID    string             `bson:"sessionId" json:"sessionId"`
	Token        string             `bson:"token" json:"token"`
	RefreshToken string             `bson:"refreshToken" json:"refreshToken"`
	ExpiresAt    primitive.DateTime `bson:"expiresAt" json:"expiresAt"`
	UserID       string             `bson:"userId" json:"userId"`
	UserAgent    string             `bson:"userAgent" json:"userAgent"`
	Device       DeviceInfo         `bson:"device" json:"device"`
	IPAddress    string             `bson:"ipAddress" json:"ipAddress"`
	Location     LocationInfo       `bson:"location" json:"location"`
	LastActive   primitive.DateTime `bson:"lastActive" json:"lastActive"`
	IssuedAt     primitive.DateTime `bson:"issuedAt" json:"issuedAt"`
	IsRevoked    bool               `bson:"isRevoked" json:"isRevoked"`
	CreatedAt    primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt    primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
}

type NewSessionArgs struct {
	UserId       string
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
	DeviceInfo   DeviceInfo
	UserAgent    string
	IpAddress    string
	Location     LocationInfo
}

func NewSession(newSessionArgs NewSessionArgs) (*Session, error) {
	now := primitive.NewDateTimeFromTime(time.Now())
	session := Session{
		SessionID:    utils.GenerateUUID(),
		Token:        newSessionArgs.Token,
		RefreshToken: newSessionArgs.RefreshToken,
		ExpiresAt:    primitive.NewDateTimeFromTime(newSessionArgs.ExpiresAt),
		UserID:       newSessionArgs.UserId,
		UserAgent:    newSessionArgs.UserAgent,
		Device:       newSessionArgs.DeviceInfo,
		IPAddress:    newSessionArgs.IpAddress,
		Location:     newSessionArgs.Location,
		LastActive:   now,
		IssuedAt:     now,
		IsRevoked:    false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := session.Validate(); err != nil {
		return nil, err
	}
	return &session, nil
}

func (session *Session) GetValue() *Session {
	return session
}

func (session *Session) Validate() error {
	validate := validator.New()
	return validate.Struct(session)
}

func (*Session) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "expiresAt", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "token", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	mongo.NewQueryBuilder[Session](db, SessionCollectionName).Query(context.Background()).CreateIndexes(indexes)
}
