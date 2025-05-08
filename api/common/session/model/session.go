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

func (session *Session) IsExpired() bool {
	return session.ExpiresAt.Time().Before(time.Now())
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

func (session *Session) GetCollectionName() string {
	return SessionCollectionName
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
		// Most critical unique indexes
		{
			Keys: bson.D{
				{Key: "token", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_session_token_unique"),
		},
		{
			Keys: bson.D{
				{Key: "sessionId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_session_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "refreshToken", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_session_refresh_token_unique"),
		},
		// Essential compound index for user session management
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "isRevoked", Value: 1},
			},
			Options: options.Index().SetName("idx_session_user_revocation"),
		},
		// TTL index for session expiry - critical for security
		{
			Keys: bson.D{
				{Key: "expiresAt", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("ttl_session_expiry"),
		},
		// TTL index for deleted sessions - 12 hours
		{
			Keys: bson.D{
				{Key: "deletedAt", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(12 * 60 * 60).SetName("ttl_session_deleted"),
		},
	}

	mongo.NewQueryBuilder[Session](db, SessionCollectionName).Query(context.Background()).CreateIndexes(indexes)
}
