package dto

import (
	"fmt"
	"sync-backend/api/moderator/model"

	"github.com/go-playground/validator/v10"
)

// CreateReportRequest represents the request to create a report
type CreateReportRequest struct {
	TargetId    string             `json:"targetId" binding:"required" validate:"required"`
	TargetType  model.ReportType   `json:"targetType" binding:"required" validate:"required,oneof=post comment user community"`
	Reason      model.ReportReason `json:"reason" binding:"required" validate:"required"`
	Description string             `json:"description,omitempty"`
	CommunityId string             `json:"communityId" binding:"required" validate:"required"`
}

func NewCreateReportRequest() *CreateReportRequest {
	return &CreateReportRequest{}
}

func (r *CreateReportRequest) GetValue() *CreateReportRequest {
	return r
}

func (r *CreateReportRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "oneof":
			if err.Field() == "TargetType" {
				msgs = append(msgs, fmt.Sprintf("%s must be one of: post, comment, user, community", err.Field()))
			}
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ProcessReportRequest represents the request to process a report
type ProcessReportRequest struct {
	ReportId       string             `uri:"reportId" binding:"required" validate:"required"`
	Status         model.ReportStatus `json:"status" binding:"required" validate:"required,oneof=approved rejected ignored"`
	ModeratorNotes string             `json:"moderatorNotes,omitempty"`
	ActionTaken    string             `json:"actionTaken,omitempty"`
}

func NewProcessReportRequest() *ProcessReportRequest {
	return &ProcessReportRequest{}
}

func (r *ProcessReportRequest) GetValue() *ProcessReportRequest {
	return r
}

func (r *ProcessReportRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "oneof":
			if err.Field() == "Status" {
				msgs = append(msgs, fmt.Sprintf("%s must be one of: approved, rejected, ignored", err.Field()))
			}
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ListReportsRequest represents the request to list reports
type ListReportsRequest struct {
	Page       int                `form:"page" query:"page" validate:"min=1"`
	Limit      int                `form:"limit" query:"limit" validate:"min=1,max=100"`
	Status     model.ReportStatus `form:"status" query:"status"`
	TargetType model.ReportType   `form:"targetType" query:"targetType"`
}

func NewListReportsRequest() *ListReportsRequest {
	return &ListReportsRequest{
		Page:  1,
		Limit: 10,
	}
}

func (r *ListReportsRequest) GetValue() *ListReportsRequest {
	return r
}

func (r *ListReportsRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s", err.Field(), err.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// GetReportRequest represents the request to get a specific report
type GetReportRequest struct {
	ReportId string `uri:"reportId" binding:"required" validate:"required"`
}

func NewGetReportRequest() *GetReportRequest {
	return &GetReportRequest{}
}

func (r *GetReportRequest) GetValue() *GetReportRequest {
	return r
}

func (r *GetReportRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
