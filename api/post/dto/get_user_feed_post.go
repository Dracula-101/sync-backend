package dto

import (
	"fmt"
	"sync-backend/api/post/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ============================================
// ||         GetUserFeedPost Request         ||
// ============================================

type GetUserFeedPostRequest struct {
	coredto.Pagination
}

func NewGetUserFeedPostRequest() *GetUserFeedPostRequest {
	return &GetUserFeedPostRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetUserFeedPostRequest) GetValue() *GetUserFeedPostRequest {
	return l
}

func (s *GetUserFeedPostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         GetUserFeedPost Response        ||
// ============================================

type GetUserFeedPostResponse struct {
	Posts []model.FeedPost `json:"posts"`
	coredto.Pagination
}

func NewGetUserFeedPostResponse(posts []model.FeedPost, page int, limit int) *GetUserFeedPostResponse {
	return &GetUserFeedPostResponse{
		Posts: posts,
		Pagination: coredto.Pagination{
			Page:  page,
			Limit: limit,
		},
	}

}

func (l *GetUserFeedPostResponse) GetValue() *GetUserFeedPostResponse {
	return l
}

func (l *GetUserFeedPostResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
