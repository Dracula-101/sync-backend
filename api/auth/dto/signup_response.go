package dto

import "github.com/go-playground/validator/v10"

type SignUpResponse struct {
	UserId string `json:"user_id" validate:"required"`
	Token  string `json:"token" validate:"required"`
}

func NewSignUpResponse(userId string, token string) *SignUpResponse {
	return &SignUpResponse{
		UserId: userId,
		Token:  token,
	}
}

func (s *SignUpResponse) GetValue() *SignUpResponse {
	return s
}

func (s *SignUpResponse) ValidateErrors(error validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range error {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, err.Field()+" is required")
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
