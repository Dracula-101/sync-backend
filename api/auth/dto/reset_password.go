package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||    Reset Password Request         ||
// =======================================

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" validate:"required,min=32"`
	NewPassword string `json:"new_password" binding:"required" validate:"required,min=8,max=100"`
}

func NewResetPasswordRequest() *ResetPasswordRequest {
	return &ResetPasswordRequest{}
}

func (r *ResetPasswordRequest) GetValue() *ResetPasswordRequest {
	return r
}

func (r *ResetPasswordRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// =======================================
// ||    Reset Password Response        ||
// =======================================

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

func NewResetPasswordResponse(message string) *ResetPasswordResponse {
	return &ResetPasswordResponse{
		Message: message,
	}
}

func (r *ResetPasswordResponse) GetValue() *ResetPasswordResponse {
	return r
}

func (r *ResetPasswordResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
