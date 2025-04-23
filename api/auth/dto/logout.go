package dto

import (
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||          Logout Request           ||
// =======================================

type LogoutRequest struct {
	coredto.BaseRequest
	UserId string `json:"user_id" binding:"required"`
}

func NewLogoutRequest() *LogoutRequest {
	return &LogoutRequest{}
}

func (l *LogoutRequest) GetValue() *LogoutRequest {
	return l
}

func (l *LogoutRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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

