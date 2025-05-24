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

const ReportCollectionName = "reports"

// ReportType defines what type of content is being reported
type ReportType string

const (
	ReportTypePost      ReportType = "post"
	ReportTypeComment   ReportType = "comment"
	ReportTypeUser      ReportType = "user"
	ReportTypeCommunity ReportType = "community"
)

// ReportReason provides standardized reasons for reporting content
type ReportReason string

const (
	ReasonSpam           ReportReason = "spam"
	ReasonHarassment     ReportReason = "harassment"
	ReasonHateSpeech     ReportReason = "hate_speech"
	ReasonNSFW           ReportReason = "nsfw_content"
	ReasonMisleading     ReportReason = "misleading"
	ReasonIllegal        ReportReason = "illegal_content"
	ReasonViolence       ReportReason = "violence"
	ReasonSelfHarm       ReportReason = "self_harm"
	ReasonMisinformation ReportReason = "misinformation"
	ReasonImpersonation  ReportReason = "impersonation"
	ReasonOther          ReportReason = "other"
)

// ReportStatus defines the status of a report in the moderation workflow
type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusApproved ReportStatus = "approved"
	ReportStatusRejected ReportStatus = "rejected"
	ReportStatusIgnored  ReportStatus = "ignored"
)

// Report represents a content report made by users
type Report struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"-"`
	ReportId       string              `bson:"reportId" json:"id"`
	CommunityId    string              `bson:"communityId" json:"communityId" validate:"required"`
	ReporterId     string              `bson:"reporterId" json:"reporterId" validate:"required"` // User who made the report
	TargetId       string              `bson:"targetId" json:"targetId" validate:"required"`     // ID of reported content
	TargetType     ReportType          `bson:"targetType" json:"targetType" validate:"required,oneof=post comment user community"`
	Reason         ReportReason        `bson:"reason" json:"reason" validate:"required"`
	Description    string              `bson:"description,omitempty" json:"description,omitempty"`
	Status         ReportStatus        `bson:"status" json:"status" validate:"required,oneof=pending approved rejected ignored"`
	ProcessedBy    string              `bson:"processedBy,omitempty" json:"processedBy,omitempty"` // Moderator who processed the report
	ProcessedAt    *primitive.DateTime `bson:"processedAt,omitempty" json:"processedAt,omitempty"`
	ModeratorNotes string              `bson:"moderatorNotes,omitempty" json:"moderatorNotes,omitempty"`
	ActionTaken    string              `bson:"actionTaken,omitempty" json:"actionTaken,omitempty"` // What action was taken if approved
	CreatedAt      primitive.DateTime  `bson:"createdAt" json:"createdAt"`
	UpdatedAt      primitive.DateTime  `bson:"updatedAt" json:"updatedAt"`
	IPAddress      string              `bson:"ipAddress,omitempty" json:"-"`
	UserAgent      string              `bson:"userAgent,omitempty" json:"-"`
}

// NewReport creates a new report
func NewReport(communityId, reporterId, targetId string, targetType ReportType, reason ReportReason) *Report {
	now := primitive.NewDateTimeFromTime(time.Now())

	return &Report{
		ReportId:    uuid.New().String(),
		CommunityId: communityId,
		ReporterId:  reporterId,
		TargetId:    targetId,
		TargetType:  targetType,
		Reason:      reason,
		Status:      ReportStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// WithDescription adds a description to the report
func (r *Report) WithDescription(description string) *Report {
	r.Description = description
	return r
}

// WithIPAddress adds IP address to the report
func (r *Report) WithIPAddress(ipAddress string) *Report {
	r.IPAddress = ipAddress
	return r
}

// WithUserAgent adds user agent to the report
func (r *Report) WithUserAgent(userAgent string) *Report {
	r.UserAgent = userAgent
	return r
}

// Process marks a report as processed
func (r *Report) Process(moderatorId string, status ReportStatus, notes string, action string) {
	now := primitive.NewDateTimeFromTime(time.Now())
	r.ProcessedBy = moderatorId
	r.ProcessedAt = &now
	r.Status = status
	r.ModeratorNotes = notes
	r.ActionTaken = action
	r.UpdatedAt = now
}

// GetValue implements mongo.Model interface
func (r *Report) GetValue() *Report {
	return r
}

// Validate implements mongo.Model interface
func (r *Report) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// GetCollectionName implements mongo.Model interface
func (r *Report) GetCollectionName() string {
	return ReportCollectionName
}

// EnsureIndexes implements mongo.Model interface
func (*Report) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "reportId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_report_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_community_status_created"),
		},
		{
			Keys: bson.D{
				{Key: "targetId", Value: 1},
				{Key: "targetType", Value: 1},
			},
			Options: options.Index().SetName("idx_target_id_type"),
		},
		{
			Keys: bson.D{
				{Key: "reporterId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_reporter_created"),
		},
	}
	mongo.NewQueryBuilder[Report](db, ReportCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
