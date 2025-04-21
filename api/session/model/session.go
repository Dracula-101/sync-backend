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
	UserID       string             `bson:"userId" json:"userId"`
	Token        string             `bson:"token" json:"token"`
	RefreshToken string             `bson:"refreshToken,omitempty" json:"refreshToken,omitempty"`
	UserAgent    string             `bson:"userAgent" json:"userAgent"`
	IPAddress    string             `bson:"ipAddress" json:"ipAddress"`
	LastActive   time.Time          `bson:"lastActive" json:"lastActive"`
	ExpiresAt    time.Time          `bson:"expiresAt" json:"expiresAt"`
	IssuedAt     time.Time          `bson:"issuedAt" json:"issuedAt"`
	IsRevoked    bool               `bson:"isRevoked" json:"isRevoked"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

func NewSession(userID, token, refreshToken, userAgent, ipAddress string, expiresAt time.Time) (*Session, error) {
	now := time.Now()
	session := Session{
		ID:           primitive.NewObjectID(),
		SessionID:    utils.GenerateUUID(),
		UserID:       userID,
		Token:        token,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		LastActive:   now,
		ExpiresAt:    expiresAt,
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
