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

const ModLogCollectionName = "moderation_logs"

// ModActionType represents the type of moderation action taken
type ModActionType string

const (
	// Content-related actions
	ActionRemovePost     ModActionType = "remove_post"
	ActionRemoveComment  ModActionType = "remove_comment"
	ActionApprovePost    ModActionType = "approve_post"
	ActionApproveComment ModActionType = "approve_comment"
	ActionPinPost        ModActionType = "pin_post"
	ActionUnpinPost      ModActionType = "unpin_post"
	ActionLockPost       ModActionType = "lock_post"
	ActionUnlockPost     ModActionType = "unlock_post"
	ActionMarkNSFW       ModActionType = "mark_nsfw"

	// User-related actions
	ActionWarnUser    ModActionType = "warn_user"
	ActionMuteUser    ModActionType = "mute_user"
	ActionUnmuteUser  ModActionType = "unmute_user"
	ActionBanUser     ModActionType = "ban_user"
	ActionTempBanUser ModActionType = "temp_ban_user"
	ActionUnbanUser   ModActionType = "unban_user"

	// Community-related actions
	ActionEditRule        ModActionType = "edit_rule"
	ActionAddRule         ModActionType = "add_rule"
	ActionRemoveRule      ModActionType = "remove_rule"
	ActionUpdateCommunity ModActionType = "update_community"
	ActionAddModerator    ModActionType = "add_moderator"
	ActionRemoveModerator ModActionType = "remove_moderator"
	ActionChangeModerator ModActionType = "change_moderator_role"

	// Report-related actions
	ActionProcessReport ModActionType = "process_report"
	ActionDismissReport ModActionType = "dismiss_report"
)

// ModLog represents a record of a moderator action
type ModLog struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	LogId         string             `bson:"logId" json:"id"`
	CommunityId   string             `bson:"communityId" json:"communityId" validate:"required"`
	ModeratorId   string             `bson:"moderatorId" json:"moderatorId" validate:"required"`
	ActionType    ModActionType      `bson:"actionType" json:"actionType" validate:"required"`
	TargetId      string             `bson:"targetId,omitempty" json:"targetId,omitempty"`     // ID of the affected post, comment, user, etc.
	TargetType    string             `bson:"targetType,omitempty" json:"targetType,omitempty"` // Type of target (post, comment, user)
	Reason        string             `bson:"reason,omitempty" json:"reason,omitempty"`
	Details       map[string]string  `bson:"details,omitempty" json:"details,omitempty"`             // Additional details about the action
	PreviousState map[string]string  `bson:"previousState,omitempty" json:"previousState,omitempty"` // Previous state for auditing
	CreatedAt     primitive.DateTime `bson:"createdAt" json:"createdAt"`
	IPAddress     string             `bson:"ipAddress,omitempty" json:"-"`
}

// NewModLog creates a new moderation log entry
func NewModLog(communityId string, moderatorId string, actionType ModActionType, targetId string, targetType string) *ModLog {
	now := primitive.NewDateTimeFromTime(time.Now())

	return &ModLog{
		LogId:       uuid.New().String(),
		CommunityId: communityId,
		ModeratorId: moderatorId,
		ActionType:  actionType,
		TargetId:    targetId,
		TargetType:  targetType,
		CreatedAt:   now,
	}
}

// WithReason adds a reason to the mod log
func (m *ModLog) WithReason(reason string) *ModLog {
	m.Reason = reason
	return m
}

// WithDetails adds details to the mod log
func (m *ModLog) WithDetails(details map[string]string) *ModLog {
	m.Details = details
	return m
}

// WithPreviousState adds previous state information to the mod log
func (m *ModLog) WithPreviousState(previousState map[string]string) *ModLog {
	m.PreviousState = previousState
	return m
}

// WithIPAddress adds IP address to the mod log
func (m *ModLog) WithIPAddress(ipAddress string) *ModLog {
	m.IPAddress = ipAddress
	return m
}

// GetValue implements mongo.Model interface
func (m *ModLog) GetValue() *ModLog {
	return m
}

// Validate implements mongo.Model interface
func (m *ModLog) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}

// GetCollectionName implements mongo.Model interface
func (m *ModLog) GetCollectionName() string {
	return ModLogCollectionName
}

// EnsureIndexes implements mongo.Model interface
func (*ModLog) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "logId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_modlog_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_community_created"),
		},
		{
			Keys: bson.D{
				{Key: "moderatorId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_moderator_created"),
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
				{Key: "actionType", Value: 1},
			},
			Options: options.Index().SetName("idx_action_type"),
		},
	}
	mongo.NewQueryBuilder[ModLog](db, ModLogCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
