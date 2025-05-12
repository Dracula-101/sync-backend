package dto

import (
	"fmt"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||       Forgot Pass Request         ||
// =======================================

type ForgotPassRequest struct {
	coredto.BaseDeviceRequest
	Email string `json:"email" binding:"required,email" validate:"required,email"`
}

func NewForgotPassRequest() *ForgotPassRequest {
	return &ForgotPassRequest{}
}

func (f *ForgotPassRequest) GetValue() *ForgotPassRequest {
	return f
}

func (f *ForgotPassRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s is not a valid email", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
