package moderatordto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// BanUserRequest represents the request to ban a user
type BanUserRequest struct {
	Reason   string `json:"reason" binding:"required" validate:"required"`
	Duration int    `json:"duration,omitempty" validate:"min=0"`
}

func NewBanUserRequest() *BanUserRequest {
	return &BanUserRequest{}
}

func (r *BanUserRequest) GetValue() *BanUserRequest {
	return r
}

func (r *BanUserRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
