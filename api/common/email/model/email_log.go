package model

import (
	"context"
	"sync-backend/arch/mongo"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const EmailLogCollectionName = "email_logs"

type EmailType string

const (
	EmailTypePasswordReset EmailType = "password_reset"
	EmailTypeVerification  EmailType = "verification"
	EmailTypeWelcome       EmailType = "welcome"
	EmailTypeNotification  EmailType = "notification"
)

type EmailStatus string

const (
	EmailStatusPending EmailStatus = "pending"
	EmailStatusSent    EmailStatus = "sent"
	EmailStatusFailed  EmailStatus = "failed"
)

type EmailLog struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	EmailId   string              `bson:"emailId" json:"emailId" validate:"required"`
	To        string              `bson:"to" json:"to" validate:"required,email"`
	Subject   string              `bson:"subject" json:"subject" validate:"required"`
	Type      EmailType           `bson:"type" json:"type" validate:"required"`
	Status    EmailStatus         `bson:"status" json:"status" validate:"required"`
	Error     string              `bson:"error,omitempty" json:"error,omitempty"`
	SentAt    *primitive.DateTime `bson:"sentAt,omitempty" json:"sentAt,omitempty"`
	CreatedAt primitive.DateTime  `bson:"createdAt" json:"createdAt"`
}

func NewEmailLog(to, subject string, emailType EmailType) *EmailLog {
	now := primitive.NewDateTimeFromTime(time.Now())
	return &EmailLog{
		EmailId:   uuid.New().String(),
		To:        to,
		Subject:   subject,
		Type:      emailType,
		Status:    EmailStatusPending,
		CreatedAt: now,
	}
}

func (e *EmailLog) GetValue() *EmailLog {
	return e
}

func (e *EmailLog) Validate() error {
	return nil
}

func (e *EmailLog) GetCollectionName() string {
	return EmailLogCollectionName
}

func (*EmailLog) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "emailId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_email_emailId_unique"),
		},
		{
			Keys: bson.D{
				{Key: "to", Value: 1},
				{Key: "type", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_email_to_type_created"),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_email_status"),
		},
		{
			Keys: bson.D{
				{Key: "createdAt", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(7776000).SetName("ttl_email_created"), // 90 days in seconds
		},
	}

	mongo.NewQueryBuilder[EmailLog](db, EmailLogCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
