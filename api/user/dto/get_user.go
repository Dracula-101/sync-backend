package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ====================================
// ||         GetUser Request         ||
// ====================================

type GetUserRequest struct {
	UserId string `uri:"userId" validate:"required"`
}

func NewGetUserRequest() *GetUserRequest {
	return &GetUserRequest{}
}

func (l *GetUserRequest) GetValue() *GetUserRequest {
	return l
}

func (s *GetUserRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
