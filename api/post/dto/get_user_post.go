package dto

import (
	"fmt"
	"sync-backend/api/post/model"

	"github.com/go-playground/validator/v10"
)

type GetUserPostRequest struct {
	Page  int `form:"page" query:"page" validate:"min=1"`
	Limit int `form:"limit" query:"limit" validate:"min=1,max=100"`
}

func NewGetUserPostRequest() *GetUserPostRequest {
	return &GetUserPostRequest{
		Page:  1,
		Limit: 10,
	}
}

func (r *GetUserPostRequest) GetValue() *GetUserPostRequest {
	return r
}

func (r *GetUserPostRequest) ValidateErrors(err validator.ValidationErrors) ([]string, error) {
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

type GetUserPostResponse struct {
	Page      int          `json:"page"`
	Limit     int          `json:"limit"`
	Total     int          `json:"total"`
	TotalPage int          `json:"total_page"`
	Posts     []model.Post `json:"posts"`
}

func NewGetUserPostResponse(posts []model.Post, page int, limit int, total int) *GetUserPostResponse {
	return &GetUserPostResponse{
		Page:      page,
		Limit:     limit,
		Total:     total,
		TotalPage: (total + limit - 1) / limit,
		Posts:     posts,
	}
}

func (r *GetUserPostResponse) GetValue() *GetUserPostResponse {
	return r
}

func (r *GetUserPostResponse) ValidateErrors(err validator.ValidationErrors) ([]string, error) {
	var errors []string
	for _, fieldErr := range err {
		switch fieldErr.Tag() {
		case "required":
			errors = append(errors, fieldErr.Field()+" is required")
		case "min":
			errors = append(errors, fieldErr.Field()+" must be greater than 0")
		default:
			errors = append(errors, fieldErr.Error())
		}
	}
	if len(errors) > 0 {
		return errors, nil
	}
	return nil, nil
}
