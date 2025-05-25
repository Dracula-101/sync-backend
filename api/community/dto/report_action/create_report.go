package reportdto

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
