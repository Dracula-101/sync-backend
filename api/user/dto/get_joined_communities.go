package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	coredto "sync-backend/arch/dto"
)

// ==============================================
// ||         JoinedCommunities Request         ||
// ==============================================

type JoinedCommunitiesRequest struct {
	coredto.Pagination
}

func NewJoinedCommunitiesRequest() *JoinedCommunitiesRequest {
	return &JoinedCommunitiesRequest{}
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
}

func NewJoinedCommunitiesResponse() *JoinedCommunitiesResponse {
	return &JoinedCommunitiesResponse{}
}

func (l *JoinedCommunitiesResponse) GetValue() *JoinedCommunitiesResponse {
	return l
}

func (l *JoinedCommunitiesResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
