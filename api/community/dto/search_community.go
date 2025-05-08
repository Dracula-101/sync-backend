package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ============================================
// ||         SearchCommunity Request         ||
// ============================================

type SearchCommunityRequest struct {
}

func NewSearchCommunityRequest() *SearchCommunityRequest {
	return &SearchCommunityRequest{}
}

func (l *SearchCommunityRequest) GetValue() *SearchCommunityRequest {
	return l
}

func (s *SearchCommunityRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
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
}

func NewSearchCommunityResponse() *SearchCommunityResponse {
	return &SearchCommunityResponse{}
}

func (l *SearchCommunityResponse) GetValue() *SearchCommunityResponse {
	return l
}

func (l *SearchCommunityResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
