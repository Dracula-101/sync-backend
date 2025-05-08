package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// =============================================
// ||         GetMyCommunities Request         ||
// =============================================

type GetMyCommunitiesRequest struct {
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
}

func NewGetMyCommunitiesResponse() *GetMyCommunitiesResponse {
	return &GetMyCommunitiesResponse{}
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
