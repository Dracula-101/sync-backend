package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ========================================
// ||         DislikePost Request         ||
// ========================================

type DislikePostRequest struct {
}

func NewDislikePostRequest() *DislikePostRequest {
	return &DislikePostRequest{}
}

func (l *DislikePostRequest) GetValue() *DislikePostRequest {
	return l
}

func (s *DislikePostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ========================================
// ||         DislikePost Response        ||
// ========================================

type DislikePostResponse struct {
	PostId     string `json:"post_id" validate:"required"`
	IsDisliked *bool  `json:"is_disliked,omitempty"`
	Synergy    *int   `json:"synergy,omitempty"`
}

func NewDislikePostResponse() *DislikePostResponse {
	return &DislikePostResponse{}
}

func (l *DislikePostResponse) GetValue() *DislikePostResponse {
	return l
}

func (l *DislikePostResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
