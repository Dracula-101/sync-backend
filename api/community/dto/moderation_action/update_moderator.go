package moderatordto

import (
	"fmt"
	"sync-backend/api/moderator/model"

	"github.com/go-playground/validator/v10"
)

// UpdateModeratorRequest represents the request to update a moderator's role or permissions
type UpdateModeratorRequest struct {
	UserId      string              `uri:"userId" binding:"required" validate:"required"`
	Role        model.ModeratorRole `json:"role,omitempty" validate:"omitempty,oneof=admin moderator content_mod user_mod auto_mod"`
	Permissions []string            `json:"permissions,omitempty"`
	Status      string              `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Notes       string              `json:"notes,omitempty"`
}

func NewUpdateModeratorRequest() *UpdateModeratorRequest {
	return &UpdateModeratorRequest{}
}

func (r *UpdateModeratorRequest) GetValue() *UpdateModeratorRequest {
	return r
}

func (r *UpdateModeratorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "oneof":
			if err.Field() == "Role" {
				msgs = append(msgs, fmt.Sprintf("%s must be one of: admin, moderator, content_mod, user_mod, auto_mod", err.Field()))
			} else if err.Field() == "Status" {
				msgs = append(msgs, fmt.Sprintf("%s must be one of: active, inactive", err.Field()))
			}
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
