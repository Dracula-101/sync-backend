package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ==================================================
// ||         UpdateUserPreferences Request         ||
// ==================================================

type UpdateUserPreferencesRequest struct {
	Language *string `form:"language" json:"language"`
	Theme    *string `form:"theme" json:"theme"`
	Timezone *string `form:"timezone" json:"timezone"`

	ShowEmailNotifications  *bool `form:"email_notifications" json:"email_notifications" validate:"omitempty,boolean"`
	ShowMobileNotifications *bool `form:"mobile_notifications" json:"mobile_notifications" validate:"omitempty,boolean"`
	ShowSensitiveContent    *bool `form:"show_sensitive_content" json:"show_sensitive_content" validate:"omitempty,boolean"`
	ShowAdultContent        *bool `form:"show_adult_content" json:"show_adult_content" validate:"omitempty,boolean"`

	IsProfileVisible           *bool `form:"is_profile_visible" json:"is_profile_visible" validate:"omitempty,boolean"`
	IsEmailVisible             *bool `form:"is_email_visible" json:"is_email_visible" validate:"omitempty,boolean"`
	IsJoinedWavelengthsVisible *bool `form:"is_joined_wavelengths_visible" json:"is_joined_wavelength_visible" validate:"omitempty,boolean"`
	FollowersVisible           *bool `form:"followers_visible" json:"follows_visible" validate:"omitempty,boolean"`
}

func NewUpdateUserPreferencesRequest() *UpdateUserPreferencesRequest {
	return &UpdateUserPreferencesRequest{}
}

func (l *UpdateUserPreferencesRequest) GetValue() *UpdateUserPreferencesRequest {
	return l
}

func (s *UpdateUserPreferencesRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "boolean":
			msgs = append(msgs, fmt.Sprintf("%s must be a boolean value", err.Field()))
		case "oneof":
			msgs = append(msgs, fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ==================================================
// ||         UpdateUserPreferences Response        ||
// ==================================================

type UpdateUserPreferencesResponse struct {
}

func NewUpdateUserPreferencesResponse() *UpdateUserPreferencesResponse {
	return &UpdateUserPreferencesResponse{}
}

func (l *UpdateUserPreferencesResponse) GetValue() *UpdateUserPreferencesResponse {
	return l
}

func (l *UpdateUserPreferencesResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
