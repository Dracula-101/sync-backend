package moderatordto

import (
	"fmt"
	"sync-backend/api/moderator/model"

	"github.com/go-playground/validator/v10"
)

// AddModeratorRequest represents the request to add a moderator
type AddModeratorRequest struct {
	UserId      string              `json:"userId" binding:"required" validate:"required"`
	Role        model.ModeratorRole `json:"role" binding:"required" validate:"required,oneof=admin moderator content_mod user_mod auto_mod"`
	Permissions []string            `json:"permissions,omitempty"`
	Notes       string              `json:"notes,omitempty"`
}

func NewAddModeratorRequest() *AddModeratorRequest {
	return &AddModeratorRequest{}
}

func (r *AddModeratorRequest) GetValue() *AddModeratorRequest {
	return r
}

func (r *AddModeratorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "oneof":
			msgs = append(msgs, fmt.Sprintf("%s must be one of: admin, moderator, content_mod, user_mod, auto_mod", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
