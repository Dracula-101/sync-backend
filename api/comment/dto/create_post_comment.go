package dto

import (
	"fmt"
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

// ==============================================
// ||         CreatePostComment Request         ||
// ==============================================

type CreatePostCommentRequest struct {
	coredto.BaseDeviceRequest
	coredto.BaseLocationRequest
	PostId      string `json:"post_id" binding:"required" validate:"required"`
	CommunityId string `json:"community_id" binding:"required" validate:"required"`
	Comment     string `json:"comment" binding:"required" validate:"required"`
	ParentId    string `json:"parent_id" binding:"omitempty" validate:"omitempty"`
}

func NewCreatePostCommentRequest() *CreatePostCommentRequest {
	return &CreatePostCommentRequest{}
}

func (l *CreatePostCommentRequest) GetValue() *CreatePostCommentRequest {
	return l
}

func (s *CreatePostCommentRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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

// ==============================================
// ||         CreatePostComment Response        ||
// ==============================================

type CreatePostCommentResponse struct {
}

func NewCreatePostCommentResponse() *CreatePostCommentResponse {
	return &CreatePostCommentResponse{}
}

func (l *CreatePostCommentResponse) GetValue() *CreatePostCommentResponse {
	return l
}

func (l *CreatePostCommentResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
