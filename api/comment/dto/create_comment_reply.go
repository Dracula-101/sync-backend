package dto

import (
	"fmt"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ===============================================
// ||         CreateCommentReply Request         ||
// ===============================================

type CreateCommentReplyRequest struct {
	coredto.BaseDeviceRequest
	coredto.BaseLocationRequest
	CommentId string `json:"commentId" validate:"required"`
	Reply     string `json:"reply" validate:"required"`
}

func NewCreateCommentReplyRequest() *CreateCommentReplyRequest {
	return &CreateCommentReplyRequest{}
}

func (l *CreateCommentReplyRequest) GetValue() *CreateCommentReplyRequest {
	return l
}

func (s *CreateCommentReplyRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         CreateCommentReply Response        ||
// ===============================================

type CreateCommentReplyResponse struct {
}

func NewCreateCommentReplyResponse() *CreateCommentReplyResponse {
	return &CreateCommentReplyResponse{}
}

func (l *CreateCommentReplyResponse) GetValue() *CreateCommentReplyResponse {
	return l
}

func (l *CreateCommentReplyResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
