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
	return &GetMyCommunitiesRequest{}
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
	NextPage    int               `json:"next_page"`
	PrevPage    int               `json:"prev_page"`
}

func NewGetMyCommunitiesResponse(communities []model.Community, total int, nextPage int, prevPage int) *GetMyCommunitiesResponse {
	return &GetMyCommunitiesResponse{
		Communities: communities,
		Total:       total,
		NextPage:    nextPage,
		PrevPage:    prevPage,
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
