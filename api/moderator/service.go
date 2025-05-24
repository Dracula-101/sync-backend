package moderator

import (
	"errors"
	"fmt"
	"sync-backend/api/moderator/model"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// moderatorService implements ModeratorService
type moderatorService struct {
	moderatorQueryBuilder mongo.QueryBuilder[model.Moderator]
	reportQueryBuilder    mongo.QueryBuilder[model.Report]
	modLogQueryBuilder    mongo.QueryBuilder[model.ModLog]
	bansQueryBuilder      mongo.QueryBuilder[model.CommunityBan]
	transactionBuilder    mongo.TransactionBuilder
}

// ModeratorService defines the interface for moderator-related operations
type ModeratorService interface {
	// Moderator management
	AddModerator(userId string, communityId string, role model.ModeratorRole, invitedBy string) (*model.Moderator, network.ApiError)
	RemoveModerator(communityId string, userId string, removedBy string, reason string) network.ApiError
	UpdateModerator(communityId string, userId string, updatedBy string, role *model.ModeratorRole, permissions []model.ModeratorPermission, status *string, notes *string) (*model.Moderator, network.ApiError)
	GetModerator(communityId string, userId string) (*model.Moderator, network.ApiError)
	ListModerators(communityId string, page, limit int) ([]*model.Moderator, int, network.ApiError)

	// Permission checks
	HasModeratorPermission(userId string, communityId string, permission model.ModeratorPermission) (bool, network.ApiError)
	IsModeratorOrHigher(userId string, communityId string) (bool, network.ApiError)
	IsAdmin(userId string, communityId string) (bool, network.ApiError)
	IsCommunityOwner(userId string, communityId string) (bool, network.ApiError)

	// User ban management
	IsUserBanned(userId string, communityId string) (bool, *model.BanInfo, network.ApiError)
	BanUser(moderatorId string, userId string, communityId string, reason string, duration *int) (*model.ModLog, network.ApiError)
	UnbanUser(moderatorId string, userId string, communityId string) (*model.ModLog, network.ApiError)

	// Reporting system
	CreateReport(reporterId string, communityId string, targetId string, targetType model.ReportType, reason model.ReportReason, description string) (*model.Report, network.ApiError)
	ProcessReport(reportId string, moderatorId string, status model.ReportStatus, notes string, action string) (*model.Report, network.ApiError)
	GetReport(reportId string) (*model.Report, network.ApiError)
	ListReports(communityId string, status *model.ReportStatus, targetType *model.ReportType, page, limit int) ([]*model.Report, int, network.ApiError)

	// Moderation actions logging
	LogModAction(communityId string, moderatorId string, actionType model.ModActionType, targetId string, targetType string, details string) (*model.ModLog, network.ApiError)
	GetModLogs(communityId string, moderatorId string, page, limit int) ([]*model.ModLog, int, network.ApiError)
}

// NewModeratorService creates a new moderator service
func NewModeratorService(
	db mongo.Database,
) ModeratorService {
	return &moderatorService{
		moderatorQueryBuilder: mongo.NewQueryBuilder[model.Moderator](db, model.ModeratorCollectionName),
		reportQueryBuilder:    mongo.NewQueryBuilder[model.Report](db, model.ReportCollectionName),
		modLogQueryBuilder:    mongo.NewQueryBuilder[model.ModLog](db, model.ModLogCollectionName),
		bansQueryBuilder:      mongo.NewQueryBuilder[model.CommunityBan](db, model.CommunityBansCollectionName),
		transactionBuilder:    mongo.NewTransactionBuilder(db),
	}
}

// AddModerator adds a new moderator to a community
func (s *moderatorService) AddModerator(userId string, communityId string, role model.ModeratorRole, invitedBy string) (*model.Moderator, network.ApiError) {
	// Check if moderator already exists
	existingMod, err := s.moderatorQueryBuilder.SingleQuery().FilterOne(
		bson.M{"userId": userId, "communityId": communityId},
		nil,
	)

	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		return nil, network.NewInternalServerError(
			"Error checking for existing moderator",
			fmt.Sprintf("Database error when checking if user '%s' is already a moderator in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	if existingMod != nil {
		return nil, network.NewConflictError(
			"Moderator already exists",
			fmt.Sprintf("User '%s' is already a moderator in community '%s'. Context - [ Duplicate Entry ]", userId, communityId),
			errors.New("moderator already exists"),
		)
	}

	// Create new moderator
	moderator := model.NewModerator(userId, communityId, role, invitedBy)

	// Insert into database
	_, err = s.moderatorQueryBuilder.SingleQuery().InsertOne(moderator)
	if err != nil {
		return nil, network.NewInternalServerError(
			"Error creating moderator",
			fmt.Sprintf("Database error when adding user '%s' as moderator in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Log the action
	_, _ = s.LogModAction(
		communityId,
		invitedBy,
		model.ActionAddModerator,
		userId,
		"user",
		fmt.Sprintf("Added user %s as %s", userId, role),
	)

	return moderator, nil
}

// RemoveModerator removes a moderator from a community
func (s *moderatorService) RemoveModerator(communityId string, userId string, removedBy string, reason string) network.ApiError {
	// Check if moderator exists
	_, err := s.moderatorQueryBuilder.SingleQuery().FilterOne(
		bson.M{"userId": userId, "communityId": communityId},
		nil,
	)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return network.NewNotFoundError(
				"Moderator not found",
				fmt.Sprintf("User '%s' is not a moderator in community '%s'. Context - [ No Data ]", userId, communityId),
				errors.New("moderator not found"),
			)
		}
		return network.NewInternalServerError(
			"Error finding moderator",
			fmt.Sprintf("Database error when finding moderator '%s' in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Delete the moderator
	_, err = s.moderatorQueryBuilder.SingleQuery().DeleteOne(
		bson.M{"userId": userId, "communityId": communityId},
		nil,
	)

	if err != nil {
		return network.NewInternalServerError(
			"Error removing moderator",
			fmt.Sprintf("Database error when removing moderator '%s' from community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Log the action
	details := fmt.Sprintf("Removed %s from moderator role", userId)
	if reason != "" {
		details += fmt.Sprintf(" - Reason: %s", reason)
	}

	_, _ = s.LogModAction(
		communityId,
		removedBy,
		model.ActionRemoveModerator,
		userId,
		"user",
		details,
	)

	return nil
}

// UpdateModerator updates a moderator's role, permissions, or status
func (s *moderatorService) UpdateModerator(communityId string, userId string, updatedBy string, role *model.ModeratorRole, permissions []model.ModeratorPermission, status *string, notes *string) (*model.Moderator, network.ApiError) {
	// Check if moderator exists
	moderator, err := s.moderatorQueryBuilder.SingleQuery().FilterOne(
		bson.M{"userId": userId, "communityId": communityId},
		nil,
	)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, network.NewNotFoundError(
				"Moderator not found",
				fmt.Sprintf("User '%s' is not a moderator in community '%s'. Context - [ No Data ]", userId, communityId),
				errors.New("moderator not found"),
			)
		}
		return nil, network.NewInternalServerError(
			"Error finding moderator",
			fmt.Sprintf("Database error when finding moderator '%s' in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	updateSet := update["$set"].(bson.M)

	if role != nil {
		updateSet["role"] = *role

		// If role changes, update permissions to match the default for that role
		if moderator.Role != *role {
			updateSet["permissions"] = model.RolePermissionMap[*role]
		}
	}

	if len(permissions) > 0 {
		updateSet["permissions"] = permissions
	}

	if status != nil {
		updateSet["status"] = *status
	}

	if notes != nil && *notes != "" {
		updateSet["notes"] = *notes
	}

	// Update the moderator
	_, err = s.moderatorQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"userId": userId, "communityId": communityId},
		update,
		nil,
	)

	if err != nil {
		return nil, network.NewInternalServerError(
			"Error updating moderator",
			fmt.Sprintf("Database error when updating moderator '%s' in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Get updated moderator
	updatedModerator, err := s.moderatorQueryBuilder.SingleQuery().FilterOne(
		bson.M{"userId": userId, "communityId": communityId},
		nil,
	)

	if err != nil {
		return nil, network.NewInternalServerError(
			"Error retrieving updated moderator",
			fmt.Sprintf("Database error when retrieving updated moderator '%s' in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Log the action
	details := "Updated moderator information"
	if role != nil {
		details += fmt.Sprintf(" - Changed role to %s", *role)
	}

	_, _ = s.LogModAction(
		communityId,
		updatedBy,
		model.ActionChangeModerator,
		userId,
		"user",
		details,
	)

	return updatedModerator, nil
}

// GetModerator gets details of a specific moderator
func (s *moderatorService) GetModerator(communityId string, userId string) (*model.Moderator, network.ApiError) {
	moderator, err := s.moderatorQueryBuilder.SingleQuery().FilterOne(
		bson.M{"userId": userId, "communityId": communityId},
		nil,
	)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, network.NewNotFoundError(
				"Moderator not found",
				fmt.Sprintf("User '%s' is not a moderator in community '%s'. Context - [ No Data ]", userId, communityId),
				errors.New("moderator not found"),
			)
		}
		return nil, network.NewInternalServerError(
			"Error finding moderator",
			fmt.Sprintf("Database error when finding moderator '%s' in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	return moderator, nil
}

// ListModerators lists all moderators for a community
func (s *moderatorService) ListModerators(communityId string, page, limit int) ([]*model.Moderator, int, network.ApiError) {
	// Calculate skip value for pagination
	skip := (page - 1) * limit

	// Get total count
	count, err := s.moderatorQueryBuilder.SingleQuery().CountDocuments(
		bson.M{"communityId": communityId},
		nil,
	)

	if err != nil {
		return nil, 0, network.NewInternalServerError(
			"Error counting moderators",
			fmt.Sprintf("Database error when counting moderators in community '%s'. Context - [ Query Failed ]", communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Get moderators with pagination
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(skip)).SetSort(bson.D{{Key: "createdAt", Value: -1}})

	moderators, err := s.moderatorQueryBuilder.SingleQuery().FindAll(
		bson.M{"communityId": communityId},
		opts,
	)

	if err != nil {
		return nil, 0, network.NewInternalServerError(
			"Error fetching moderators",
			fmt.Sprintf("Database error when fetching moderators in community '%s'. Context - [ Query Failed ]", communityId),
			network.DB_ERROR,
			err,
		)
	}

	return moderators, int(count), nil
}

// HasModeratorPermission checks if a user has a specific permission in a community
func (s *moderatorService) HasModeratorPermission(userId string, communityId string, permission model.ModeratorPermission) (bool, network.ApiError) {
	// Get the moderator document
	moderator, err := s.moderatorQueryBuilder.SingleQuery().FilterOne(
		bson.M{"userId": userId, "communityId": communityId, "status": "active"},
		nil,
	)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return false, nil // Not a moderator, so no permission
		}
		return false, network.NewInternalServerError(
			"Error checking for moderator permission",
			fmt.Sprintf("Database error when checking permission for user '%s' in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Check if the moderator has the permission
	for _, p := range moderator.Permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// IsModeratorOrHigher checks if the user has any moderator role in the community
func (s *moderatorService) IsModeratorOrHigher(userId string, communityId string) (bool, network.ApiError) {
	// Check if the user has a moderator document
	count, err := s.moderatorQueryBuilder.SingleQuery().CountDocuments(
		bson.M{"userId": userId, "communityId": communityId, "status": "active"},
		nil,
	)

	if err != nil {
		return false, network.NewInternalServerError(
			"Error checking if user is moderator",
			fmt.Sprintf("Database error when checking if user '%s' is a moderator in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	return count > 0, nil
}

// IsAdmin checks if the user has an admin role in the community
func (s *moderatorService) IsAdmin(userId string, communityId string) (bool, network.ApiError) {
	// Check if the user has an admin role
	count, err := s.moderatorQueryBuilder.SingleQuery().CountDocuments(
		bson.M{
			"userId":      userId,
			"communityId": communityId,
			"role":        model.RoleAdmin,
			"status":      "active",
		},
		nil,
	)

	if err != nil {
		return false, network.NewInternalServerError(
			"Error checking if user is admin",
			fmt.Sprintf("Database error when checking if user '%s' is an admin in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	return count > 0, nil
}

// IsCommunityOwner checks if the user is the owner of the community
func (s *moderatorService) IsCommunityOwner(userId string, communityId string) (bool, network.ApiError) {
	// Check if the user is the owner
	count, err := s.moderatorQueryBuilder.SingleQuery().CountDocuments(
		bson.M{
			"userId":      userId,
			"communityId": communityId,
			"role":        model.RoleOwner,
			"status":      "active",
		},
		nil,
	)

	if err != nil {
		return false, network.NewInternalServerError(
			"Error checking if user is community owner",
			fmt.Sprintf("Database error when checking if user '%s' is the owner of community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	return count > 0, nil
}

// IsUserBanned checks if a user is banned from a community
func (s *moderatorService) IsUserBanned(userId string, communityId string) (bool, *model.BanInfo, network.ApiError) {
	ban, err := s.bansQueryBuilder.SingleQuery().FilterOne(
		bson.M{
			"userId":      userId,
			"communityId": communityId,
			"isActive":    true,
		},
		nil,
	)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return false, nil, nil // User is not banned
		}
		return false, nil, network.NewInternalServerError(
			"Error checking if user is banned",
			fmt.Sprintf("Database error when checking if user '%s' is banned from community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Check if temporary ban has expired
	if ban.ExpiresAt != nil {
		expiryTime := ban.ExpiresAt.Time()
		if time.Now().After(expiryTime) {
			// Ban has expired, update it
			_, err := s.bansQueryBuilder.SingleQuery().UpdateOne(
				bson.M{"_id": ban.ID},
				bson.M{"$set": bson.M{"isActive": false}},
				nil,
			)

			if err != nil {
				// Log error but continue with considering user not banned
				return false, nil, nil
			}

			return false, nil, nil // Ban expired
		}
	}

	// User is banned
	info := &model.BanInfo{
		Reason:      ban.Reason,
		IsPermanent: ban.ExpiresAt == nil,
	}

	if ban.ExpiresAt != nil {
		expTime := ban.ExpiresAt.Time()
		info.ExpiresAt = &expTime
	}

	return true, info, nil
}

// BanUser bans a user from a community
func (s *moderatorService) BanUser(moderatorId string, userId string, communityId string, reason string, duration *int) (*model.ModLog, network.ApiError) {
	now := time.Now()
	ptNow := primitive.NewDateTimeFromTime(now)

	ban := &model.CommunityBan{
		ID:          primitive.NewObjectID(),
		BanId:       primitive.NewObjectID().Hex(),
		CommunityId: communityId,
		UserId:      userId,
		ModeratorId: moderatorId,
		Reason:      reason,
		IsActive:    true,
		CreatedAt:   ptNow,
		UpdatedAt:   ptNow,
	}

	// Set expiration if duration is provided
	if duration != nil && *duration > 0 {
		expires := now.Add(time.Duration(*duration) * 24 * time.Hour) // Duration in days
		ptExpires := primitive.NewDateTimeFromTime(expires)
		ban.ExpiresAt = &ptExpires
		ban.Duration = duration
	}

	// Create or update ban
	_, err := s.bansQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"userId": userId, "communityId": communityId},
		bson.M{"$set": ban},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		return nil, network.NewInternalServerError(
			"Error banning user",
			fmt.Sprintf("Database error when banning user '%s' from community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Log the action
	actionType := model.ActionBanUser
	if duration != nil && *duration > 0 {
		actionType = model.ActionTempBanUser
	}

	details := fmt.Sprintf("Banned user %s", userId)
	if duration != nil && *duration > 0 {
		details += fmt.Sprintf(" for %d days", *duration)
	}
	if reason != "" {
		details += fmt.Sprintf(" - Reason: %s", reason)
	}

	log, err := s.LogModAction(
		communityId,
		moderatorId,
		actionType,
		userId,
		"user",
		details,
	)

	if err != nil {
		return nil, network.NewInternalServerError(
			"Error logging ban action",
			fmt.Sprintf("Database error when logging ban action for user '%s' in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	return log, nil
}

// UnbanUser unbans a user from a community
func (s *moderatorService) UnbanUser(moderatorId string, userId string, communityId string) (*model.ModLog, network.ApiError) {
	// Find and update the ban
	_, err := s.bansQueryBuilder.SingleQuery().UpdateOne(
		bson.M{
			"userId":      userId,
			"communityId": communityId,
			"isActive":    true,
		},
		bson.M{
			"$set": bson.M{
				"isActive":  false,
				"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
			},
		},
		nil,
	)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, network.NewNotFoundError(
				"Ban not found",
				fmt.Sprintf("User '%s' is not banned in community '%s'. Context - [ No Data ]", userId, communityId),
				errors.New("ban not found"),
			)
		}
		return nil, network.NewInternalServerError(
			"Error unbanning user",
			fmt.Sprintf("Database error when unbanning user '%s' from community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Log the action
	details := fmt.Sprintf("Unbanned user %s", userId)

	log, err := s.LogModAction(
		communityId,
		moderatorId,
		model.ActionUnbanUser,
		userId,
		"user",
		details,
	)

	if err != nil {
		return nil, network.NewInternalServerError(
			"Error logging unban action",
			fmt.Sprintf("Database error when logging unban action for user '%s' in community '%s'. Context - [ Query Failed ]", userId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	return log, nil
}

// CreateReport creates a new report
func (s *moderatorService) CreateReport(reporterId string, communityId string, targetId string, targetType model.ReportType, reason model.ReportReason, description string) (*model.Report, network.ApiError) {
	report := model.NewReport(reporterId, communityId, targetId, targetType, reason).WithDescription(description)

	_, err := s.reportQueryBuilder.SingleQuery().InsertOne(report)
	if err != nil {
		return nil, network.NewInternalServerError(
			"Error creating report",
			fmt.Sprintf("Database error when creating report for '%s' in community '%s'. Context - [ Query Failed ]", targetId, communityId),
			network.DB_ERROR,
			err,
		)
	}

	return report, nil
}

// ProcessReport processes a report and updates its status
func (s *moderatorService) ProcessReport(reportId string, moderatorId string, status model.ReportStatus, notes string, action string) (*model.Report, network.ApiError) {
	// Find the report
	report, err := s.reportQueryBuilder.SingleQuery().FilterOne(
		bson.M{"reportId": reportId},
		nil,
	)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, network.NewNotFoundError(
				"Report not found",
				fmt.Sprintf("Report with ID '%s' not found. Context - [ No Data ]", reportId),
				errors.New("report not found"),
			)
		}
		return nil, network.NewInternalServerError(
			"Error finding report",
			fmt.Sprintf("Database error when finding report with ID '%s'. Context - [ Query Failed ]", reportId),
			network.DB_ERROR,
			err,
		)
	}

	// Update report
	update := bson.M{
		"$set": bson.M{
			"status":           status,
			"processedBy":      moderatorId,
			"processedAt":      primitive.NewDateTimeFromTime(time.Now()),
			"processingNotes":  notes,
			"moderationAction": action,
		},
	}

	_, err = s.reportQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"reportId": reportId},
		update,
		nil,
	)

	if err != nil {
		return nil, network.NewInternalServerError(
			"Error updating report",
			fmt.Sprintf("Database error when updating report with ID '%s'. Context - [ Query Failed ]", reportId),
			network.DB_ERROR,
			err,
		)
	}

	// Get updated report
	updatedReport, err := s.reportQueryBuilder.SingleQuery().FilterOne(
		bson.M{"reportId": reportId},
		nil,
	)

	if err != nil {
		return nil, network.NewInternalServerError(
			"Error retrieving updated report",
			fmt.Sprintf("Database error when retrieving updated report with ID '%s'. Context - [ Query Failed ]", reportId),
			network.DB_ERROR,
			err,
		)
	}

	// Log the action
	details := fmt.Sprintf("Processed report %s - Status: %s", reportId, status)
	if action != "" {
		details += fmt.Sprintf(", Action: %s", action)
	}

	_, _ = s.LogModAction(
		report.CommunityId,
		moderatorId,
		model.ActionProcessReport,
		reportId,
		"report",
		details,
	)

	return updatedReport, nil
}

// GetReport gets details of a specific report
func (s *moderatorService) GetReport(reportId string) (*model.Report, network.ApiError) {
	report, err := s.reportQueryBuilder.SingleQuery().FilterOne(
		bson.M{"reportId": reportId},
		nil,
	)

	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, network.NewNotFoundError(
				"Report not found",
				fmt.Sprintf("Report with ID '%s' not found. Context - [ No Data ]", reportId),
				errors.New("report not found"),
			)
		}
		return nil, network.NewInternalServerError(
			"Error finding report",
			fmt.Sprintf("Database error when finding report with ID '%s'. Context - [ Query Failed ]", reportId),
			network.DB_ERROR,
			err,
		)
	}

	return report, nil
}

// ListReports lists reports for a community
func (s *moderatorService) ListReports(communityId string, status *model.ReportStatus, targetType *model.ReportType, page, limit int) ([]*model.Report, int, network.ApiError) {
	// Calculate skip value for pagination
	skip := (page - 1) * limit

	// Build filter
	filter := bson.M{"communityId": communityId}

	if status != nil {
		filter["status"] = *status
	}

	if targetType != nil {
		filter["targetType"] = *targetType
	}

	// Get total count
	count, err := s.reportQueryBuilder.SingleQuery().CountDocuments(filter, nil)
	if err != nil {
		return nil, 0, network.NewInternalServerError(
			"Error counting reports",
			fmt.Sprintf("Database error when counting reports in community '%s'. Context - [ Query Failed ]", communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Get reports with pagination
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(skip)).SetSort(bson.D{{Key: "createdAt", Value: -1}})

	reports, err := s.reportQueryBuilder.SingleQuery().FindAll(filter, opts)
	if err != nil {
		return nil, 0, network.NewInternalServerError(
			"Error fetching reports",
			fmt.Sprintf("Database error when fetching reports in community '%s'. Context - [ Query Failed ]", communityId),
			network.DB_ERROR,
			err,
		)
	}

	return reports, int(count), nil
}

// LogModAction logs a moderation action
func (s *moderatorService) LogModAction(communityId string, moderatorId string, actionType model.ModActionType, targetId string, targetType string, details string) (*model.ModLog, network.ApiError) {
	modLog := model.NewModLog(communityId, moderatorId, actionType, targetId, targetType).WithDetails(map[string]string{"details": details})

	_, err := s.modLogQueryBuilder.SingleQuery().InsertOne(modLog)
	if err != nil {
		return nil, network.NewInternalServerError(
			"Error logging moderation action",
			fmt.Sprintf("Database error when logging moderation action in community '%s'. Context - [ Query Failed ]", communityId),
			network.DB_ERROR,
			err,
		)
	}

	return modLog, nil
}

// GetModLogs gets moderation logs for a community
func (s *moderatorService) GetModLogs(communityId string, moderatorId string, page, limit int) ([]*model.ModLog, int, network.ApiError) {
	// Calculate skip value for pagination
	skip := (page - 1) * limit

	// Build filter
	filter := bson.M{"communityId": communityId}
	if moderatorId != "" {
		filter["moderatorId"] = moderatorId
	}

	// Get total count
	count, err := s.modLogQueryBuilder.SingleQuery().CountDocuments(filter, nil)
	if err != nil {
		return nil, 0, network.NewInternalServerError(
			"Error counting moderation logs",
			fmt.Sprintf("Database error when counting moderation logs in community '%s'. Context - [ Query Failed ]", communityId),
			network.DB_ERROR,
			err,
		)
	}

	// Get logs with pagination
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(skip)).SetSort(bson.D{{Key: "timestamp", Value: -1}})

	logs, err := s.modLogQueryBuilder.SingleQuery().FindAll(filter, opts)
	if err != nil {
		return nil, 0, network.NewInternalServerError(
			"Error fetching moderation logs",
			fmt.Sprintf("Database error when fetching moderation logs in community '%s'. Context - [ Query Failed ]", communityId),
			network.DB_ERROR,
			err,
		)
	}

	return logs, int(count), nil
}
