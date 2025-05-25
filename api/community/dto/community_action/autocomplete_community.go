package communitydto

import (
	"fmt"
	community "sync-backend/api/community/model"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ==================================================
// ||         AutocompleteCommunity Request         ||
// ==================================================

type AutocompleteCommunityRequest struct {
	Query       string `form:"query" validate:"required"`
	ShowPrivate bool   `form:"show_private" query:"show_private"`
	coredto.Pagination
}

func NewAutocompleteCommunityRequest() *AutocompleteCommunityRequest {
	return &AutocompleteCommunityRequest{
		Pagination: *coredto.NewPagination(),
	}
}

func (l *AutocompleteCommunityRequest) GetValue() *AutocompleteCommunityRequest {
	return l
}

func (s *AutocompleteCommunityRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s", err.Field(), err.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ==================================================
// ||         AutocompleteCommunity Response        ||
// ==================================================

type AutocompleteCommunityResponse struct {
	Communities []community.CommunityAutocomplete `json:"communities"`
	Total       int                               `json:"total"`
	NextPage    int                               `json:"next_page"`
	PrevPage    int                               `json:"prev_page"`
	HasNext     bool                              `json:"has_next"`
	HasPrev     bool                              `json:"has_prev"`
	CurrentPage int                               `json:"current_page"`
	Limit       int                               `json:"limit"`
	TotalCount  int                               `json:"total_count"`
}

func NewAutocompleteCommunityResponse(communities []community.CommunityAutocomplete, total int, nextPage int, prevPage int, hasNext bool, hasPrev bool, currentPage int, limit int, totalCount int) *AutocompleteCommunityResponse {
	return &AutocompleteCommunityResponse{
		Communities: communities,
		Total:       total,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
		CurrentPage: currentPage,
		Limit:       limit,
		TotalCount:  totalCount,
	}
}

func (l *AutocompleteCommunityResponse) GetValue() *AutocompleteCommunityResponse {
	return l
}

func (l *AutocompleteCommunityResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
