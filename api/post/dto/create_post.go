package dto

import (
	"sync-backend/api/post/model"

	"github.com/go-playground/validator/v10"
)

// =======================================
// ||       Create Post Request          ||
// =======================================

type CreatePostRequest struct {
	Title       string         `form:"title" json:"title" validate:"required,min=1,max=300"`
	Content     string         `form:"content" json:"content" validate:"required"`
	Tags        []string       `form:"tags" json:"tags"`
	Media       []string       `form:"media" json:"media"`
	Type        model.PostType `form:"type" json:"type" validate:"required,oneof=text image video link"`
	CommunityId string         `form:"communityId" json:"communityId" validate:"required"`
	IsNSFW      bool           `form:"isNSFW,omitempty" json:"isNSFW"`
	IsSpoiler   bool           `form:"isSpoiler,omitempty" json:"isSpoiler"`
}

func NewCreatePostRequest() *CreatePostRequest {
	return &CreatePostRequest{
		Tags:  []string{},
		Media: []string{},
	}
}

func (r *CreatePostRequest) GetValue() *CreatePostRequest {
	return r
}

func (r *CreatePostRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, err.Field()+" is required")
		case "min":
			msgs = append(msgs, err.Field()+" must be at least "+err.Param()+" characters")
		case "max":
			msgs = append(msgs, err.Field()+" must be at most "+err.Param()+" characters")
		case "oneof":
			msgs = append(msgs, err.Field()+" must be one of "+err.Param())
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil

}

// =======================================
// ||       Create Post Response         ||
// =======================================

type CreatePostResponse struct {
	PostId string `json:"postId"`
}

func NewCreatePostResponse(postId string) *CreatePostResponse {
	return &CreatePostResponse{
		PostId: postId,
	}
}

func (r *CreatePostResponse) GetValue() *CreatePostResponse {
	return r
}

func (r *CreatePostResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
