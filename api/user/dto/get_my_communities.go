package dto

import (
	"fmt"
	"sync-backend/api/community/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// =============================================
// ||         GetMyCommunities Request         ||
// =============================================

type GetMyCommunitiesRequest struct {
	coredto.Pagination
}

func NewGetMyCommunitiesRequest() *GetMyCommunitiesRequest {
	return &GetMyCommunitiesRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *GetMyCommunitiesRequest) GetValue() *GetMyCommunitiesRequest {
	return l
}

func (s *GetMyCommunitiesRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
// ||         GetMyCommunities Response        ||
// =============================================

type GetMyCommunitiesResponse struct {
	Communities []model.Community `json:"communities"`
	Total       int               `json:"total"`
}

func NewGetMyCommunitiesResponse(communities []model.Community, total int) *GetMyCommunitiesResponse {
	return &GetMyCommunitiesResponse{
		Communities: communities,
		Total:       total,
	}
}

func (l *GetMyCommunitiesResponse) GetValue() *GetMyCommunitiesResponse {
	return l
}

func (l *GetMyCommunitiesResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
