package moderatordto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// RemoveModeratorRequest represents the request to remove a moderator
type RemoveModeratorRequest struct {
	UserId string `uri:"userId" binding:"required" validate:"required"`
	Reason string `json:"reason,omitempty"`
}

func NewRemoveModeratorRequest() *RemoveModeratorRequest {
	return &RemoveModeratorRequest{}
}

func (r *RemoveModeratorRequest) GetValue() *RemoveModeratorRequest {
	return r
}

func (r *RemoveModeratorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
