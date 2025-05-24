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

const ModeratorCollectionName = "moderators"

// ModeratorRole defines the role of a moderator with corresponding permissions
type ModeratorRole string

const (
	// Admin has all permissions and can manage other moderators
	RoleAdmin ModeratorRole = "admin"
	// Moderator can moderate content, users, and apply basic moderation actions
	RoleModerator ModeratorRole = "moderator"
	// ContentMod can only moderate content (posts, comments)
	RoleContentMod ModeratorRole = "content_mod"
	// UserMod can only moderate users (warnings, temporary bans)
	RoleUserMod ModeratorRole = "user_mod"
	// AutoMod is a special role for bot/automated moderation
	RoleAutoMod ModeratorRole = "auto_mod"
)

// ModeratorPermission defines specific permissions a moderator can have
type ModeratorPermission string

const (
	// Content moderation permissions
	PermissionRemoveContent  ModeratorPermission = "remove_content"
	PermissionApproveContent ModeratorPermission = "approve_content"
	PermissionPinContent     ModeratorPermission = "pin_content"
	PermissionLockContent    ModeratorPermission = "lock_content"
	PermissionMarkNSFW       ModeratorPermission = "mark_nsfw"

	// User moderation permissions
	PermissionWarnUser  ModeratorPermission = "warn_user"
	PermissionMuteUser  ModeratorPermission = "mute_user"
	PermissionBanUser   ModeratorPermission = "ban_user"
	PermissionUnbanUser ModeratorPermission = "unban_user"

	// Community permissions
	PermissionEditRules        ModeratorPermission = "edit_rules"
	PermissionEditCommunity    ModeratorPermission = "edit_community"
	PermissionManageFlairs     ModeratorPermission = "manage_flairs"
	PermissionManageModerators ModeratorPermission = "manage_moderators"
	PermissionViewReports      ModeratorPermission = "view_reports"
	PermissionProcessReports   ModeratorPermission = "process_reports"
	PermissionViewModLog       ModeratorPermission = "view_mod_log"
	PermissionViewAnalytics    ModeratorPermission = "view_analytics"
)

// RolePermissionMap maps moderator roles to their default permissions
var RolePermissionMap = map[ModeratorRole][]ModeratorPermission{
	RoleAdmin: {
		PermissionRemoveContent, PermissionApproveContent, PermissionPinContent,
		PermissionLockContent, PermissionMarkNSFW, PermissionWarnUser,
		PermissionMuteUser, PermissionBanUser, PermissionUnbanUser,
		PermissionEditRules, PermissionEditCommunity, PermissionManageFlairs,
		PermissionManageModerators, PermissionViewReports, PermissionProcessReports,
		PermissionViewModLog, PermissionViewAnalytics,
	},
	RoleModerator: {
		PermissionRemoveContent, PermissionApproveContent, PermissionPinContent,
		PermissionLockContent, PermissionMarkNSFW, PermissionWarnUser,
		PermissionMuteUser, PermissionBanUser, PermissionUnbanUser,
		PermissionEditRules, PermissionViewReports, PermissionProcessReports,
		PermissionViewModLog,
	},
	RoleContentMod: {
		PermissionRemoveContent, PermissionApproveContent, PermissionPinContent,
		PermissionLockContent, PermissionMarkNSFW, PermissionViewReports,
		PermissionProcessReports, PermissionViewModLog,
	},
	RoleUserMod: {
		PermissionWarnUser, PermissionMuteUser, PermissionViewReports,
		PermissionProcessReports, PermissionViewModLog,
	},
	RoleAutoMod: {
		PermissionRemoveContent, PermissionMarkNSFW, PermissionViewReports,
	},
}

// Moderator represents a user with moderation privileges in a community
type Moderator struct {
	ID          primitive.ObjectID    `bson:"_id,omitempty" json:"-"`
	ModeratorId string                `bson:"moderatorId" json:"id"`
	UserId      string                `bson:"userId" json:"userId" validate:"required"`
	CommunityId string                `bson:"communityId" json:"communityId" validate:"required"`
	Role        ModeratorRole         `bson:"role" json:"role" validate:"required,oneof=admin moderator content_mod user_mod auto_mod"`
	Permissions []ModeratorPermission `bson:"permissions" json:"permissions"`
	InvitedBy   string                `bson:"invitedBy,omitempty" json:"invitedBy,omitempty"`
	InvitedAt   primitive.DateTime    `bson:"invitedAt" json:"invitedAt"`
	Status      string                `bson:"status" json:"status" validate:"required,oneof=active inactive pending"`
	Notes       string                `bson:"notes,omitempty" json:"notes,omitempty"`
	Stats       ModeratorStats        `bson:"stats" json:"stats"`
	CreatedAt   primitive.DateTime    `bson:"createdAt" json:"createdAt"`
	UpdatedAt   primitive.DateTime    `bson:"updatedAt" json:"updatedAt"`
	DeletedAt   *primitive.DateTime   `bson:"deletedAt,omitempty" json:"-"`
}

// ModeratorStats tracks moderator actions and activity
type ModeratorStats struct {
	ContentRemoved   int64              `bson:"contentRemoved" json:"contentRemoved"`
	ContentApproved  int64              `bson:"contentApproved" json:"contentApproved"`
	UsersWarned      int64              `bson:"usersWarned" json:"usersWarned"`
	UsersMuted       int64              `bson:"usersMuted" json:"usersMuted"`
	UsersBanned      int64              `bson:"usersBanned" json:"usersBanned"`
	ReportsProcessed int64              `bson:"reportsProcessed" json:"reportsProcessed"`
	LastActiveAt     primitive.DateTime `bson:"lastActiveAt" json:"lastActiveAt"`
}

// NewModerator creates a new moderator for a community
func NewModerator(userId string, communityId string, role ModeratorRole, invitedBy string) *Moderator {
	now := primitive.NewDateTimeFromTime(time.Now())

	// Get default permissions for the role
	permissions := RolePermissionMap[role]

	return &Moderator{
		ModeratorId: uuid.New().String(),
		UserId:      userId,
		CommunityId: communityId,
		Role:        role,
		Permissions: permissions,
		InvitedBy:   invitedBy,
		InvitedAt:   now,
		Status:      "active",
		Stats: ModeratorStats{
			ContentRemoved:   0,
			ContentApproved:  0,
			UsersWarned:      0,
			UsersMuted:       0,
			UsersBanned:      0,
			ReportsProcessed: 0,
			LastActiveAt:     now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// HasPermission checks if the moderator has a specific permission
func (m *Moderator) HasPermission(permission ModeratorPermission) bool {
	for _, p := range m.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// AddPermission adds a permission to the moderator if they don't already have it
func (m *Moderator) AddPermission(permission ModeratorPermission) {
	if !m.HasPermission(permission) {
		m.Permissions = append(m.Permissions, permission)
	}
}

// RemovePermission removes a permission from the moderator
func (m *Moderator) RemovePermission(permission ModeratorPermission) {
	var newPermissions []ModeratorPermission
	for _, p := range m.Permissions {
		if p != permission {
			newPermissions = append(newPermissions, p)
		}
	}
	m.Permissions = newPermissions
}

// GetValue implements mongo.Model interface
func (m *Moderator) GetValue() *Moderator {
	return m
}

// Validate implements mongo.Model interface
func (m *Moderator) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}

// GetCollectionName implements mongo.Model interface
func (m *Moderator) GetCollectionName() string {
	return ModeratorCollectionName
}

// EnsureIndexes implements mongo.Model interface
func (*Moderator) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "moderatorId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_moderator_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "communityId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_user_community_unique"),
		},
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
				{Key: "role", Value: 1},
			},
			Options: options.Index().SetName("idx_community_role"),
		},
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_community_status"),
		},
		{
			Keys: bson.D{
				{Key: "deletedAt", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(30 * 24 * 60 * 60).SetName("ttl_moderator_deleted"),
		},
	}
	mongo.NewQueryBuilder[Moderator](db, ModeratorCollectionName).Query(context.Background()).CheckIndexes(indexes)
}
