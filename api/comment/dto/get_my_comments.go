package dto

import (
	"fmt"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ==========================================
// ||         GetMyComments Request         ||
// ==========================================

type GetMyCommentsRequest struct {
	coredto.Pagination
}

func NewGetMyCommentsRequest() *GetMyCommentsRequest {
	return &GetMyCommentsRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetMyCommentsRequest) GetValue() *GetMyCommentsRequest {
	return l
}

func (s *GetMyCommentsRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         GetMyComments Response        ||
// ==========================================

type GetMyCommentsResponse struct {
}

func NewGetMyCommentsResponse() *GetMyCommentsResponse {
	return &GetMyCommentsResponse{}
}

func (l *GetMyCommentsResponse) GetValue() *GetMyCommentsResponse {
	return l
}

func (l *GetMyCommentsResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
