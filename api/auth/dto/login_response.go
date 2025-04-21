package dto

import "github.com/go-playground/validator/v10"

type LoginResponse struct {
	UserId      string `json:"user_id" validate:"required"`
	SessionId   string `json:"session_id" validate:"required"`
	AccessToken string `json:"access_token" validate:"required"`
}

func NewLoginResponse(userId string, sessionId string, accessToken string) *LoginResponse {
	return &LoginResponse{
		UserId:      userId,
		SessionId:   sessionId,
		AccessToken: accessToken,
	}
}

func (l *LoginResponse) GetValue() *LoginResponse {
	return l
}

func (l *LoginResponse) ValidateErrors(error validator.ValidationErrors) ([]string, error) {
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
