package dto

import (
	"fmt"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ===================================================
// ||         GetTrendingCommunities Request         ||
// ===================================================

type GetTrendingCommunitiesRequest struct {
	coredto.Pagination
}

func NewGetTrendingCommunitiesRequest() *GetTrendingCommunitiesRequest {
	return &GetTrendingCommunitiesRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetTrendingCommunitiesRequest) GetValue() *GetTrendingCommunitiesRequest {
	return l
}

func (s *GetTrendingCommunitiesRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ===================================================
// ||         GetTrendingCommunities Response        ||
// ===================================================

type GetTrendingCommunitiesResponse struct {
}

func NewGetTrendingCommunitiesResponse() *GetTrendingCommunitiesResponse {
	return &GetTrendingCommunitiesResponse{}
}

func (l *GetTrendingCommunitiesResponse) GetValue() *GetTrendingCommunitiesResponse {
	return l
}

func (l *GetTrendingCommunitiesResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
