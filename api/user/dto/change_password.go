package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ===========================================
// ||         ChangePassword Request         ||
// ===========================================

type ChangePasswordRequest struct {
	OldPassword string `form:"old_password" json:"old_password" binding:"required,min=6,max=100"`
	NewPassword string `form:"new_password" json:"new_password" binding:"required,min=6,max=100"`
}

func NewChangePasswordRequest() *ChangePasswordRequest {
	return &ChangePasswordRequest{}
}

func (l *ChangePasswordRequest) GetValue() *ChangePasswordRequest {
	return l
}

func (s *ChangePasswordRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ===========================================
// ||         ChangePassword Response        ||
// ===========================================

type ChangePasswordResponse struct {
}

func NewChangePasswordResponse() *ChangePasswordResponse {
	return &ChangePasswordResponse{}
}

func (l *ChangePasswordResponse) GetValue() *ChangePasswordResponse {
	return l
}

func (l *ChangePasswordResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
