package dto

import (
	"fmt"
	"sync-backend/api/comment/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ==========================================
// ||         GetPostComment Request         ||
// ==========================================

type GetPostCommentRequest struct {
	coredto.Pagination
}

func NewGetPostComentRequest() *GetPostCommentRequest {
	return &GetPostCommentRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetPostCommentRequest) GetValue() *GetPostCommentRequest {
	return l
}

func (s *GetPostCommentRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ==========================================
// ||         GetPostComment Response        ||
// ==========================================

type GetPostCommentResponse struct {
	Comments []model.Comment `json:"comments"`
	Total    int             `json:"total"`
}

func NewGetPostComentResponse() *GetPostCommentResponse {
	return &GetPostCommentResponse{}
}

func (l *GetPostCommentResponse) GetValue() *GetPostCommentResponse {
	return l
}

func (l *GetPostCommentResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
