package dto

import (
	"fmt"
	"sync-backend/api/post/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// =============================================
// ||         GetUserSavedPost Request         ||
// =============================================

type GetUserSavedPostRequest struct {
	coredto.Pagination
}

func NewGetUserSavedPostRequest() *GetUserSavedPostRequest {
	return &GetUserSavedPostRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetUserSavedPostRequest) GetValue() *GetUserSavedPostRequest {
	return l
}

func (s *GetUserSavedPostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         GetUserSavedPost Response        ||
// =============================================

type GetUserSavedPostResponse struct {
	Posts []model.FeedPost `json:"posts"`
	coredto.Pagination
}

func NewGetUserSavedPostResponse(posts []model.FeedPost, page int, limit int) *GetUserSavedPostResponse {
	return &GetUserSavedPostResponse{
		Posts:      posts,
		Pagination: coredto.Pagination{Page: page, Limit: limit},
	}
}

func (l *GetUserSavedPostResponse) GetValue() *GetUserSavedPostResponse {
	return l
}

func (l *GetUserSavedPostResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
