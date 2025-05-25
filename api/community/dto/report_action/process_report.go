package reportdto

import (
	"fmt"
	"sync-backend/api/moderator/model"

	"github.com/go-playground/validator/v10"
)

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