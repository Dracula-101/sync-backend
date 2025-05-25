package communitydto

import "github.com/go-playground/validator/v10"

type JoinCommunityRequest struct {
	CommunityId string `uri:"communityId" binding:"required" validate:"required"`
}

func NewJoinCommunityRequest() *JoinCommunityRequest {
	return &JoinCommunityRequest{}
}

func (r *JoinCommunityRequest) GetValue() *JoinCommunityRequest {
	return r
}

func (r *JoinCommunityRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
