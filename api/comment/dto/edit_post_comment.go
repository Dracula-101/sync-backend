package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ============================================
// ||         EditPostComment Request         ||
// ============================================

type EditPostCommentRequest struct {
	Comment  string `json:"comment" binding:"required" validate:"required"`
	ParentId string `json:"parent_id" binding:"omitempty" validate:"omitempty"`
}

func NewEditPostCommentRequest() *EditPostCommentRequest {
	return &EditPostCommentRequest{}
}

func (l *EditPostCommentRequest) GetValue() *EditPostCommentRequest {
	return l
}

func (s *EditPostCommentRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         EditPostComment Response        ||
// ============================================

type EditPostCommentResponse struct {
}

func NewEditPostCommentResponse() *EditPostCommentResponse {
	return &EditPostCommentResponse{}
}

func (l *EditPostCommentResponse) GetValue() *EditPostCommentResponse {
	return l
}

func (l *EditPostCommentResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
