package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type SignUpRequest struct {
	Name          string `json:"name" binding:"required" validate:"required,max=200"`
	Email         string `json:"email" binding:"required,email" validate:"email"`
	Password      string `json:"password" binding:"required" validate:"required,min=6,max=100"`
	ProfilePicUrl string `json:"profile_pic_url" binding:"omitempty,max=500" validate:"omitempty,max=500"`
	DeviceId      string `json:"device_id"`
	UserAgent     string `json:"user_agent"`
	IPAddress     string `json:"ip_address"`
	DeviceName    string `json:"device_name"`
}

func NewSignUpRequest() *SignUpRequest {
	return &SignUpRequest{}
}

func (s *SignUpRequest) GetValue() *SignUpRequest {
	return s
}

func (s *SignUpRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param()))
		case "name":
			msgs = append(msgs, fmt.Sprintf("%s must be a valid name", err.Field()))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s is not a valid email", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
