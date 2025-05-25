package reportdto

import (
	"fmt"
	"sync-backend/api/moderator/model"

	"github.com/go-playground/validator/v10"
)

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