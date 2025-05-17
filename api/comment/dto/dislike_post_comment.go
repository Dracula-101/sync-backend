package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ===============================================
// ||         DislikePostComment Request         ||
// ===============================================

type DislikePostCommentRequest struct {
}

func NewDislikePostCommentRequest() *DislikePostCommentRequest {
	return &DislikePostCommentRequest{}
}

func (l *DislikePostCommentRequest) GetValue() *DislikePostCommentRequest {
	return l
}

func (s *DislikePostCommentRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ===============================================
// ||         DislikePostComment Response        ||
// ===============================================

type DislikePostCommentResponse struct {
	IsDisliked bool `json:"isDisliked"`
	Synergy    int  `json:"synergy"`
}

func NewDislikePostCommentResponse(isDisliked bool, synergy int) *DislikePostCommentResponse {
	return &DislikePostCommentResponse{
		IsDisliked: isDisliked,
		Synergy:    synergy,
	}
}

func (l *DislikePostCommentResponse) GetValue() *DislikePostCommentResponse {
	return l
}

func (l *DislikePostCommentResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
