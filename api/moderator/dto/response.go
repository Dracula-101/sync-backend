package dto

import (
	"sync-backend/api/moderator/model"
)

// ListModeratorsResponse is the response for listing moderators
type ListModeratorsResponse struct {
	Moderators []*model.Moderator `json:"moderators"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int                `json:"total"`
}

// NewListModeratorsResponse creates a new response for listing moderators
func NewListModeratorsResponse(moderators []*model.Moderator, page, limit, total int) *ListModeratorsResponse {
	return &ListModeratorsResponse{
		Moderators: moderators,
		Page:       page,
		Limit:      limit,
		Total:      total,
	}
}

// ListReportsResponse is the response for listing reports
type ListReportsResponse struct {
	Reports []*model.Report `json:"reports"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	Total   int             `json:"total"`
}

// NewListReportsResponse creates a new response for listing reports
func NewListReportsResponse(reports []*model.Report, page, limit, total int) *ListReportsResponse {
	return &ListReportsResponse{
		Reports: reports,
		Page:    page,
		Limit:   limit,
		Total:   total,
	}
}

// ListModLogsResponse is the response for listing moderation logs
type ListModLogsResponse struct {
	ModLogs []*model.ModLog `json:"modLogs"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	Total   int             `json:"total"`
}

// NewListModLogsResponse creates a new response for listing moderation logs
func NewListModLogsResponse(modLogs []*model.ModLog, page, limit, total int) *ListModLogsResponse {
	return &ListModLogsResponse{
		ModLogs: modLogs,
		Page:    page,
		Limit:   limit,
		Total:   total,
	}
}

// PermissionCheckResponse is the response for checking a permission
type PermissionCheckResponse struct {
	HasPermission bool `json:"hasPermission"`
}

// NewPermissionCheckResponse creates a new response for checking a permission
func NewPermissionCheckResponse(hasPermission bool) *PermissionCheckResponse {
	return &PermissionCheckResponse{
		HasPermission: hasPermission,
	}
}
