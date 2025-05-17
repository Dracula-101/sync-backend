package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ======================================
// ||         Like Post Request         ||
// ======================================

type LikePostRequest struct {
}

func NewLikePostRequest() *LikePostRequest {
	return &LikePostRequest{}
}

func (l *LikePostRequest) GetValue() *LikePostRequest {
	return l
}

func (s *LikePostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ======================================
// ||         Like Post Response        ||
// ======================================

type LikePostResponse struct {
	PostId  string `json:"post_id" validate:"required"`
	IsLiked *bool  `json:"is_liked,omitempty"`
	Synergy *int   `json:"synergy,omitempty"`
}

func NewLikePostResponse() *LikePostResponse {
	return &LikePostResponse{}
}

func (l *LikePostResponse) GetValue() *LikePostResponse {
	return l
}

func (l *LikePostResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
