package dto

import (
	"fmt"

	"sync-backend/api/user/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||          Signup Request           ||
// =======================================

type SignUpRequest struct {
	coredto.BaseDeviceRequest
	coredto.BaseLocationRequest
	UserName      string `json:"username" binding:"required" validate:"required,min=3,max=50"`
	Email         string `json:"email" binding:"required,email" validate:"email"`
	Password      string `json:"password" binding:"required" validate:"required,min=6,max=100"`
	ProfilePicUrl string `json:"profile_pic_url" binding:"omitempty,max=500" validate:"omitempty,max=500"`
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
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s is not a valid email", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// =======================================
// ||          Signup Response           ||
// =======================================

type SignUpResponse struct {
	User         model.UserInfo `json:"user"`
	AccessToken  string         `json:"access_token" validate:"required"`
	RefreshToken string         `json:"refresh_token" validate:"required"`
}

func NewSignUpResponse(userInfo model.UserInfo, accessToken string, refreshToken string) *SignUpResponse {
	return &SignUpResponse{
		User:         userInfo,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func (s *SignUpResponse) GetValue() *SignUpResponse {
	return s
}

func (s *SignUpResponse) ValidateErrors(error validator.ValidationErrors) ([]string, error) {
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
