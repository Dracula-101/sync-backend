package moderatordto

import (
	"fmt"
	"sync-backend/api/moderator/model"

	"github.com/go-playground/validator/v10"
)

// ListModeratorsRequest represents the request to list moderators of a community
type ListModeratorsRequest struct {
	Page  int `form:"page" query:"page" validate:"min=1"`
	Limit int `form:"limit" query:"limit" validate:"min=1,max=100"`
}

func NewListModeratorsRequest() *ListModeratorsRequest {
	return &ListModeratorsRequest{
		Page:  1,
		Limit: 10,
	}
}

func (r *ListModeratorsRequest) GetValue() *ListModeratorsRequest {
	return r
}

func (r *ListModeratorsRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
