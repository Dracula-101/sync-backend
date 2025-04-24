package model

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type UserInfo struct {
	UserId     string         `json:"id" validate:"required"`
	Email      string         `json:"email"`
	Name       string         `json:"name"`
	ProfilePic string         `json:"profile_pic"`
	Providers  []ProviderInfo `json:"provider,omitempty"`
}

func (u *UserInfo) GetValue() *UserInfo {
	return u
}

func (u *UserInfo) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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

type ProviderInfo struct {
	ProviderName string    `json:"name"`
	AddedAt      time.Time `json:"addedAt"`
}
