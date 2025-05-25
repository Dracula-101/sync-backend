package reportdto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

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
