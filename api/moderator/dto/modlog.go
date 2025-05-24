package dto

import (
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ListModLogsRequest is the request for listing moderation logs
type ListModLogsRequest struct {
	coredto.Pagination
	ModeratorId string `form:"moderatorId" json:"moderatorId"`
}

// NewListModLogsRequest creates a new request for listing moderation logs
func NewListModLogsRequest() *ListModLogsRequest {
	return &ListModLogsRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *ListModLogsRequest) GetValue() *ListModLogsRequest {
	return l
}

func (l *ListModLogsRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, err.Field()+" is required")
		case "oneof":
			msgs = append(msgs, err.Field()+" must be one of: post, comment, user, community")
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
