package moderatordto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// GetModeratorRequest represents the request to get a specific moderator
type GetModeratorRequest struct {
	UserId string `uri:"userId" binding:"required" validate:"required"`
}

func NewGetModeratorRequest() *GetModeratorRequest {
	return &GetModeratorRequest{}
}

func (r *GetModeratorRequest) GetValue() *GetModeratorRequest {
	return r
}

func (r *GetModeratorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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