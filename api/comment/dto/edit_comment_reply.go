package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// =============================================
// ||         EditCommentReply Request         ||
// =============================================

type EditCommentReplyRequest struct {
	CommentId string `json:"comment_id" binding:"required" validate:"required"`
	Reply     string `json:"reply" binding:"required" validate:"required"`
}

func NewEditCommentReplyRequest() *EditCommentReplyRequest {
	return &EditCommentReplyRequest{}
}

func (l *EditCommentReplyRequest) GetValue() *EditCommentReplyRequest {
	return l
}

func (s *EditCommentReplyRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// =============================================
// ||         EditCommentReply Response        ||
// =============================================

type EditCommentReplyResponse struct {
}

func NewEditCommentReplyResponse() *EditCommentReplyResponse {
	return &EditCommentReplyResponse{}
}

func (l *EditCommentReplyResponse) GetValue() *EditCommentReplyResponse {
	return l
}

func (l *EditCommentReplyResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
