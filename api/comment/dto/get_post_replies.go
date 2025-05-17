package dto

import (
	"fmt"

	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ===========================================
// ||         GetPostReplies Request         ||
// ===========================================

type GetPostRepliesParams struct {
	coredto.Pagination
}

func NewGetPostRepliesParams() *GetPostRepliesParams {
	return &GetPostRepliesParams{
		Pagination: *coredto.NewPagination(),
	}
}

type GetPostRepliesRequest struct {
	PostId string `json:"post_id" binding:"required" validate:"required"`
}

func NewGetPostRepliesRequest() *GetPostRepliesRequest {
	return &GetPostRepliesRequest{}
}

func (l *GetPostRepliesRequest) GetValue() *GetPostRepliesRequest {
	return l
}

func (s *GetPostRepliesRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         GetPostReplies Response        ||
// ===========================================

type GetPostRepliesResponse struct {
}

func NewGetPostRepliesResponse() *GetPostRepliesResponse {
	return &GetPostRepliesResponse{}
}

func (l *GetPostRepliesResponse) GetValue() *GetPostRepliesResponse {
	return l
}

func (l *GetPostRepliesResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
