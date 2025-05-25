package moderatordto

import (
	"sync-backend/api/moderator/model"
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
