package dto

import (
	"fmt"
	"sync-backend/api/post/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ============================================
// ||         GetTrendingPost Request         ||
// ============================================

type GetTrendingPostRequest struct {
	coredto.Pagination
}

func NewGetTrendingPostRequest() *GetTrendingPostRequest {
	return &GetTrendingPostRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetTrendingPostRequest) GetValue() *GetTrendingPostRequest {
	return l
}

func (s *GetTrendingPostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         GetTrendingPost Response        ||
// ============================================

type GetTrendingPostResponse struct {
	Posts []model.FeedPost `json:"posts"`
	coredto.Pagination
}

func NewGetTrendingPostResponse(posts []model.FeedPost, page int, limit int) *GetTrendingPostResponse {
	return &GetTrendingPostResponse{
		Posts:      posts,
		Pagination: coredto.Pagination{page, limit},
	}
}

func (l *GetTrendingPostResponse) GetValue() *GetTrendingPostResponse {
	return l
}

func (l *GetTrendingPostResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
