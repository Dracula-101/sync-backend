package dto

import (
	"fmt"
	"sync-backend/api/moderator/model"

	"github.com/go-playground/validator/v10"
)

// AddModeratorRequest represents the request to add a moderator
type AddModeratorRequest struct {
	UserId      string              `json:"userId" binding:"required" validate:"required"`
	Role        model.ModeratorRole `json:"role" binding:"required" validate:"required,oneof=admin moderator content_mod user_mod auto_mod"`
	Permissions []string            `json:"permissions,omitempty"`
	Notes       string              `json:"notes,omitempty"`
}

func NewAddModeratorRequest() *AddModeratorRequest {
	return &AddModeratorRequest{}
}

func (r *AddModeratorRequest) GetValue() *AddModeratorRequest {
	return r
}

func (r *AddModeratorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "oneof":
			msgs = append(msgs, fmt.Sprintf("%s must be one of: admin, moderator, content_mod, user_mod, auto_mod", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// RemoveModeratorRequest represents the request to remove a moderator
type RemoveModeratorRequest struct {
	UserId string `uri:"userId" binding:"required" validate:"required"`
	Reason string `json:"reason,omitempty"`
}

func NewRemoveModeratorRequest() *RemoveModeratorRequest {
	return &RemoveModeratorRequest{}
}

func (r *RemoveModeratorRequest) GetValue() *RemoveModeratorRequest {
	return r
}

func (r *RemoveModeratorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// UpdateModeratorRequest represents the request to update a moderator's role or permissions
type UpdateModeratorRequest struct {
	UserId      string              `uri:"userId" binding:"required" validate:"required"`
	Role        model.ModeratorRole `json:"role,omitempty" validate:"omitempty,oneof=admin moderator content_mod user_mod auto_mod"`
	Permissions []string            `json:"permissions,omitempty"`
	Status      string              `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Notes       string              `json:"notes,omitempty"`
}

func NewUpdateModeratorRequest() *UpdateModeratorRequest {
	return &UpdateModeratorRequest{}
}

func (r *UpdateModeratorRequest) GetValue() *UpdateModeratorRequest {
	return r
}

func (r *UpdateModeratorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		case "oneof":
			if err.Field() == "Role" {
				msgs = append(msgs, fmt.Sprintf("%s must be one of: admin, moderator, content_mod, user_mod, auto_mod", err.Field()))
			} else if err.Field() == "Status" {
				msgs = append(msgs, fmt.Sprintf("%s must be one of: active, inactive", err.Field()))
			}
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// GetModeratorRequest represents the request to get a specific moderator
type GetModeratorRequest struct {
	UserId string `uri:"userId" binding:"required" validate:"required"`
}

func NewGetModeratorRequest() *GetModeratorRequest {
	return &GetModeratorRequest{}
}

func (r *GetModeratorRequest) GetValue() *GetModeratorRequest {
	return r
}

func (r *GetModeratorRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

// ListModeratorsRequest represents the request to list moderators of a community
type ListModeratorsRequest struct {
	Page  int `form:"page" query:"page" validate:"min=1"`
	Limit int `form:"limit" query:"limit" validate:"min=1,max=100"`
}

func NewListModeratorsRequest() *ListModeratorsRequest {
	return &ListModeratorsRequest{
		Page:  1,
		Limit: 10,
	}
}

func (r *ListModeratorsRequest) GetValue() *ListModeratorsRequest {
	return r
}

func (r *ListModeratorsRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s", err.Field(), err.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
