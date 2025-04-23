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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
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

// =======================================
// ||           Login Response           ||
// =======================================

type LoginResponse struct {
	User        model.UserInfo `json:"user"`
	AccessToken string   `json:"access_token" validate:"required"`
}

func NewLoginResponse(userInfo model.UserInfo, accessToken string) *LoginResponse {
	return &LoginResponse{
		User:        userInfo,
		AccessToken: accessToken,
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
