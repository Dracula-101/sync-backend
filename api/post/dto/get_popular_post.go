package dto

import (
	"fmt"
	"sync-backend/api/post/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ===========================================
// ||         GetPopularPost Request         ||
// ===========================================

type GetPopularPostRequest struct {
	coredto.Pagination
}

func NewGetPopularPostRequest() *GetPopularPostRequest {
	return &GetPopularPostRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetPopularPostRequest) GetValue() *GetPopularPostRequest {
	return l
}

func (s *GetPopularPostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         GetPopularPost Response        ||
// ===========================================

type GetPopularPostResponse struct {
	Posts []model.FeedPost `json:"posts"`
	coredto.Pagination
}

func NewGetPopularPostResponse(posts []model.FeedPost, page int, limit int) *GetPopularPostResponse {
	return &GetPopularPostResponse{
		Posts:      posts,
		Pagination: coredto.Pagination{Page: page, Limit: limit},
	}
}

func (l *GetPopularPostResponse) GetValue() *GetPopularPostResponse {
	return l
}

func (l *GetPopularPostResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
