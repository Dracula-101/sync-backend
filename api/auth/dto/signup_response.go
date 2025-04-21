package dto

import "github.com/go-playground/validator/v10"

type SignUpResponse struct {
	UserId       string `json:"user_id" validate:"required"`
	SessiondId   string `json:"session_id" validate:"required"`
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func NewSignUpResponse(userId string, sessionId string, accessToken, refreshToken string) *SignUpResponse {
	return &SignUpResponse{
		UserId:       userId,
		SessiondId:   sessionId,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
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
