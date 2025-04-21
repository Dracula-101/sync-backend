package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceId   string `json:"device_id" binding:"omitempty,max=100"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	DeviceName string `json:"device_name"`
}

func NewLoginRequest() *LoginRequest {
	return &LoginRequest{}
}

func (l *LoginRequest) GetValue() *LoginRequest {
	return l
}

func (l *LoginRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
