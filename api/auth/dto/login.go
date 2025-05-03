package dto

import (
	"fmt"
	"sync-backend/api/user/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||           Login Request           ||
// =======================================

type LoginRequest struct {
	coredto.BaseRequest
	Email    string `json:"email" binding:"required,email" validate:"email"`
	Password string `json:"password" binding:"required" validate:"required,min=6,max=100"`
}

func NewLoginRequest() *LoginRequest {
	return &LoginRequest{}
}

func (l *LoginRequest) GetValue() *LoginRequest {
	return l
}

func (s *LoginRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param()))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s is not a valid email", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// =======================================
// ||           Login Response           ||
// =======================================

type LoginResponse struct {
	User        model.UserInfo `json:"user"`
	AccessToken string         `json:"access_token" validate:"required"`
	RefreshToken string         `json:"refresh_token" validate:"required"`
}

func NewLoginResponse(userInfo model.UserInfo, accessToken string, refreshToken string) *LoginResponse {
	return &LoginResponse{
		User:        userInfo,
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}
}

func (l *LoginResponse) GetValue() *LoginResponse {
	return l
}

func (l *LoginResponse) ValidateErrors(error validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range error {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, err.Field()+" is required")
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
