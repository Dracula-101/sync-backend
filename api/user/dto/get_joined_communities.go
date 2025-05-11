package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"sync-backend/api/community/model"
	coredto "sync-backend/arch/dto"
)

// ==============================================
// ||         JoinedCommunities Request         ||
// ==============================================

type JoinedCommunitiesRequest struct {
	coredto.Pagination
	
}

func NewJoinedCommunitiesRequest() *JoinedCommunitiesRequest {
	return &JoinedCommunitiesRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *JoinedCommunitiesRequest) GetValue() *JoinedCommunitiesRequest {
	return l
}

func (s *JoinedCommunitiesRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ==============================================
// ||         JoinedCommunities Response        ||
// ==============================================

type JoinedCommunitiesResponse struct {
	Communities []model.Community `json:"communities"`
	Total       int               `json:"total"`
}

func NewJoinedCommunitiesResponse(communities []model.Community, total int) *JoinedCommunitiesResponse {
	return &JoinedCommunitiesResponse{
		Communities: communities,
		Total:       total,
	}
}
