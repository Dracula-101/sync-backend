package dto

import (
	"fmt"
	"sync-backend/api/user/model"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||    Verify Email Request           ||
// =======================================

type VerifyEmailRequest struct {
	Token string `uri:"token" binding:"required" validate:"required,min=32"`
}

func NewVerifyEmailRequest() *VerifyEmailRequest {
	return &VerifyEmailRequest{}
}

func (v *VerifyEmailRequest) GetValue() *VerifyEmailRequest {
	return v
}

func (v *VerifyEmailRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// =======================================
// ||    Verify Email Response          ||
// =======================================

type VerifyEmailResponse struct {
	Message string         `json:"message"`
	User    model.UserInfo `json:"user"`
}

func NewVerifyEmailResponse(message string, user model.UserInfo) *VerifyEmailResponse {
	return &VerifyEmailResponse{
		Message: message,
		User:    user,
	}
}

func (v *VerifyEmailResponse) GetValue() *VerifyEmailResponse {
	return v
}

func (v *VerifyEmailResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, err.Field()+" is required")
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
