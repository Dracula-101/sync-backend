package dto

import (
	"fmt"
	"sync-backend/api/post/model"

	"github.com/go-playground/validator/v10"
)

type GetCommunityPostRequest struct {
	Page  int `form:"page" query:"page" validate:"min=1"`
	Limit int `form:"limit" query:"limit" validate:"min=1,max=100"`
}

func NewGetCommunityPostRequest() *GetCommunityPostRequest {
	return &GetCommunityPostRequest{
		Page:  1,
		Limit: 10,
	}
}

func (r *GetCommunityPostRequest) GetValue() *GetCommunityPostRequest {
	return r
}

func (r *GetCommunityPostRequest) ValidateErrors(err validator.ValidationErrors) ([]string, error) {
	var errors []string
	for _, fieldErr := range err {
		switch fieldErr.Tag() {
		case "required":
			errors = append(errors, fieldErr.Field()+" is required")
		case "min":
			errors = append(errors, fmt.Sprintf("%s must be greater than %s", fieldErr.Field(), fieldErr.Param()))
		case "max":
			errors = append(errors, fmt.Sprintf("%s must be less than %s", fieldErr.Field(), fieldErr.Param()))
		default:
			errors = append(errors, fieldErr.Error())
		}
	}
	if len(errors) > 0 {
		return errors, nil
	}
	return nil, nil
}

type GetCommunityPostResponse struct {
	Posts      []model.Post `json:"posts"`
	TotalPosts int          `json:"totalPosts"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
}

func NewGetCommunityPostResponse(posts []model.Post, page int, limit int, totalPosts int) *GetCommunityPostResponse {
	return &GetCommunityPostResponse{
		Posts:      posts,
		Page:       page,
		Limit:      limit,
		TotalPosts: totalPosts,
	}
}
