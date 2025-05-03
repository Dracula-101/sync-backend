package dto

import "github.com/go-playground/validator/v10"

type LeaveCommunityRequest struct {
	CommunityId string `uri:"communityId" binding:"required" validate:"required"`
}

func NewLeaveCommunityRequest() *LeaveCommunityRequest {
	return &LeaveCommunityRequest{}
}

func (r *LeaveCommunityRequest) GetValue() *LeaveCommunityRequest {
	return r
}

func (r *LeaveCommunityRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, err.Field()+" is required")
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
