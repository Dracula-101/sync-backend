package dto

import (
	"fmt"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||         UpdateUser Request         ||
// =======================================

type UpdateUserRequest struct {
	Bio                string                 `form:"bio" json:"bio" binding:"omitempty" validate:"omitempty,max=500"`
	Avatar             []multipart.FileHeader `form:"avatar" json:"avatar" binding:"omitempty" validate:"omitempty,dive"`
	Background         []multipart.FileHeader `form:"background" json:"background" binding:"omitempty" validate:"omitempty,dive"`
	AvatarFilePath     *string
	BackgroundFilePath *string
}

func NewUpdateUserRequest() *UpdateUserRequest {
	return &UpdateUserRequest{}
}

func (l *UpdateUserRequest) GetValue() *UpdateUserRequest {
	return l
}

func (s *UpdateUserRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s exceeds maximum length of %s characters", err.Field(), err.Param()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
		case "dive":
			msgs = append(msgs, fmt.Sprintf("%s must be a valid file", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// =======================================
// ||         UpdateUser Response        ||
// =======================================

type UpdateUserResponse struct {
}

func NewUpdateUserResponse() *UpdateUserResponse {
	return &UpdateUserResponse{}
}

func (l *UpdateUserResponse) GetValue() *UpdateUserResponse {
	return l
}

func (l *UpdateUserResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
