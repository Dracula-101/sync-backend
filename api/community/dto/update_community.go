package dto

import (
	"fmt"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
)

// ============================================
// ||         UpdateCommunity Request         ||
// ============================================

type UpdateCommunityRequest struct {
	CommunityDescription string                  `form:"description" json:"description" binding:"omitempty"`
	AvatarPhoto          *[]multipart.FileHeader `form:"avatar_photo" json:"avatar_photo" binding:"omitempty"`
	BackgroundPhoto      *[]multipart.FileHeader `form:"background_photo" json:"background_photo" binding:"omitempty"`
	AvatarFilePath       string
	BackgroundFilePath   string
}

func NewUpdateCommunityRequest() *UpdateCommunityRequest {
	return &UpdateCommunityRequest{}
}

func (l *UpdateCommunityRequest) GetValue() *UpdateCommunityRequest {
	return l
}

func (s *UpdateCommunityRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ============================================
// ||         UpdateCommunity Response        ||
// ============================================

type UpdateCommunityResponse struct {
}

func NewUpdateCommunityResponse() *UpdateCommunityResponse {
	return &UpdateCommunityResponse{}
}

func (l *UpdateCommunityResponse) GetValue() *UpdateCommunityResponse {
	return l
}

func (l *UpdateCommunityResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
