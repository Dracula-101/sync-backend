package dto

import "github.com/go-playground/validator/v10"

type TogglePostResponse struct {
	PostId string `json:"postId"`
	Field  string `json:"field"`
	Value  bool   `json:"value"`
}

func NewTogglePostResponse(postId, field string, value bool) *TogglePostResponse {
	return &TogglePostResponse{
		PostId: postId,
		Field:  field,
		Value:  value,
	}
}

func (r *TogglePostResponse) GetValue() *TogglePostResponse {
	return r
}

func (r *TogglePostResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		msgs = append(msgs, err.Field()+" is invalid")
	}
	return msgs, nil
}
