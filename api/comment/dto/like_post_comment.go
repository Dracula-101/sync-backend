package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ============================================
// ||         LikePostComment Request         ||
// ============================================

type LikePostCommentRequest struct {
}

func NewLikePostCommentRequest() *LikePostCommentRequest {
	return &LikePostCommentRequest{}
}

func (l *LikePostCommentRequest) GetValue() *LikePostCommentRequest {
	return l
}

func (s *LikePostCommentRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ============================================
// ||         LikePostComment Response        ||
// ============================================

type LikePostCommentResponse struct {
	IsLiked bool `json:"is_liked"`
	Synergy int  `json:"synergy"`
}

func NewLikePostCommentResponse(isLiked bool, synergy int) *LikePostCommentResponse {
	return &LikePostCommentResponse{
		IsLiked: isLiked,
		Synergy: synergy,
	}
}

func (l *LikePostCommentResponse) GetValue() *LikePostCommentResponse {
	return l
}

func (l *LikePostCommentResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
