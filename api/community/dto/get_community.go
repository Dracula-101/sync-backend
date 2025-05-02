package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ========================================
// ||       Get Community Request         ||
// ========================================

type GetCommunityRequest struct {
	Id string `uri:"communityId" binding:"required" validate:"required"`
}

func NewGetCommunityRequest() *GetCommunityRequest {
	return &GetCommunityRequest{}
}

func (c *GetCommunityRequest) GetValue() *GetCommunityRequest {
	return c
}

func (c *GetCommunityRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
