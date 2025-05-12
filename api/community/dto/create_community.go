package dto

import (
	"fmt"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
)

// ==========================================
// ||       Create Community Request        ||
// ==========================================

type CreateCommunityRequest struct {
	Name               string                `form:"name" binding:"required" validate:"required,min=3,max=50"`
	Description        string                `form:"description" binding:"required" validate:"required,min=15,max=500"`
	TagIds             []string              `form:"tag_ids" binding:"required" validate:"required"`
	AvatarPhoto        *multipart.FileHeader `form:"avatar_photo" binding:"omitempty" validate:"omitempty"`
	BackgroundPhoto    *multipart.FileHeader `form:"background_photo" binding:"omitempty" validate:"omitempty"`
	AvatarFilePath     string
	BackgroundFilePath string
}

func NewCreateCommunityRequest() *CreateCommunityRequest {
	return &CreateCommunityRequest{}
}

func (c *CreateCommunityRequest) GetValue() *CreateCommunityRequest {
	return c
}

func (c *CreateCommunityRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ==========================================
// ||       Create Community Response       ||
// ==========================================

type CreateCommunityResponse struct {
	CommunityId string `json:"community_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
}

func NewCreateCommunityResponse(communityId string, name string, slug string) *CreateCommunityResponse {
	return &CreateCommunityResponse{
		CommunityId: communityId,
		Name:        name,
		Slug:        slug,
	}
}

func (c *CreateCommunityResponse) GetValue() *CreateCommunityResponse {
	return c
}

func (c *CreateCommunityResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
