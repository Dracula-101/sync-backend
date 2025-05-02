package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ==========================================
// ||       Create Community Request        ||
// ==========================================

type CreateCommunityRequest struct {
	Name          string   `json:"name" binding:"required" validate:"required,min=3,max=50"`
	Description   string   `json:"description" binding:"required" validate:"required,min=15,max=500"`
	TagIds        []string `json:"tag_ids" binding:"required" validate:"required"`
	AvatarUrl     string   `json:"avatar_url" binding:"omitempty" validate:"omitempty"`
	BackgroundUrl string   `json:"background_url" binding:"omitempty" validate:"omitempty"`
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
