package dto

import (
	"fmt"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ===========================================
// ||         GetUserComment Request         ||
// ===========================================

type GetUserCommentRequest struct {
	coredto.Pagination
}

func NewGetUserCommentRequest() *GetUserCommentRequest {
	return &GetUserCommentRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetUserCommentRequest) GetValue() *GetUserCommentRequest {
	return l
}

func (s *GetUserCommentRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ===========================================
// ||         GetUserComment Response        ||
// ===========================================

type GetUserCommentResponse struct {
}

func NewGetUserCommentResponse() *GetUserCommentResponse {
	return &GetUserCommentResponse{}
}

func (l *GetUserCommentResponse) GetValue() *GetUserCommentResponse {
	return l
}

func (l *GetUserCommentResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
