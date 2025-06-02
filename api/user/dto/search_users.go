package dto

import (
	"fmt"
	"sync-backend/api/user/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ========================================
// ||         SearchUsers Request         ||
// ========================================

type SearchUsersRequest struct {
	coredto.Pagination
	Query string `form:"query" json:"query" binding:"required,min=3,max=100"`
}

func NewSearchUsersRequest() *SearchUsersRequest {
	return &SearchUsersRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *SearchUsersRequest) GetValue() *SearchUsersRequest {
	return l
}

func (s *SearchUsersRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ========================================
// ||         SearchUsers Response        ||
// ========================================

type SearchUsersResponse struct {
	Users []model.SearchUser `json:"users" validate:"required,dive"`
	Total int64              `json:"total" validate:"required"`
	Page  int64              `json:"page" validate:"required"`
	Limit int64              `json:"limit" validate:"required"`
}

func NewSearchUsersResponse(users []model.SearchUser, total, page, limit int64) *SearchUsersResponse {
	return &SearchUsersResponse{
		Users: users,
		Total: total,
		Page:  page,
		Limit: limit,
	}
}

func (l *SearchUsersResponse) GetValue() *SearchUsersResponse {
	return l
}

func (l *SearchUsersResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
