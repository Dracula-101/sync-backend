package dto

import (
	"sync-backend/api/user/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||          Google Request           ||
// =======================================


type GoogleLoginRequest struct {
	coredto.BaseDeviceRequest
	coredto.BaseLocationRequest
	GoogleIdToken string `json:"google_id_token" binding:"required" validate:"required"`
	Username      string `json:"username" binding:"required" validate:"required,min=3,max=50"`
}

func NewGoogleLoginRequest() *GoogleLoginRequest {
	return &GoogleLoginRequest{}
}

func (l *GoogleLoginRequest) GetValue() *GoogleLoginRequest {
	return l
}

func (l *GoogleLoginRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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

// =======================================
// ||          Google Response           ||
// =======================================

type GoogleLoginResponse struct {
	User         model.UserInfo `json:"user"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
}

func NewGoogleLoginResponse(userInfo model.UserInfo, accessToken string, refreshToken string) *GoogleLoginResponse {
	return &GoogleLoginResponse{
		User:         userInfo,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func (g *GoogleLoginResponse) GetValue() *GoogleLoginResponse {
	return g
}

func (g *GoogleLoginResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
