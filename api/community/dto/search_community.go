package dto

import (
	"fmt"
	"sync-backend/api/community/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ============================================
// ||         SearchCommunity Request         ||
// ============================================

type SearchCommunityRequest struct {
	Query       string `form:"query" query:"query" validate:"required"`
	ShowPrivate bool   `form:"show_private" query:"show_private"`
	coredto.Pagination
}

func NewSearchCommunityRequest() *SearchCommunityRequest {
	return &SearchCommunityRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *SearchCommunityRequest) GetValue() *SearchCommunityRequest {
	return l
}

func (s *SearchCommunityRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ============================================
// ||         SearchCommunity Response        ||
// ============================================

type SearchCommunityResponse struct {
	Communities []model.CommunitySearchResult `json:"communities"`
	Total       int                           `json:"total"`
	NextPage    int                           `json:"next_page"`
	PrevPage    int                           `json:"prev_page"`
	HasNext     bool                          `json:"has_next"`
	HasPrev     bool                          `json:"has_prev"`
	CurrentPage int                           `json:"current_page"`
	Limit       int                           `json:"limit"`
	TotalCount  int                           `json:"total_count"`
}
