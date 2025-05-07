package dto

import (
	"sync-backend/api/post/model"

	"github.com/go-playground/validator/v10"
)

type EditPostRequest struct {
	Title       string         `json:"title" form:"title"`
	Content     string         `json:"content" form:"content"`
	PostType    model.PostType `json:"postType" form:"postType" validate:"oneof=TEXT IMAGE VIDEO"`
	CommunityId string         `json:"communityId" form:"communityId"`
	IsNSFW      bool           `json:"isNSFW" form:"isNSFW"`
	IsSpoiler   bool           `json:"isSpoiler" form:"isSpoiler"`
}

func NewEditPostRequest() *EditPostRequest {
	return &EditPostRequest{}
}

func (r *EditPostRequest) GetValue() *EditPostRequest {
	return r
}

func (r *EditPostRequest) ValidateErrors(err validator.ValidationErrors) ([]string, error) {
	var errors []string
	for _, fieldErr := range err {
		switch fieldErr.Tag() {
		case "oneof":
			errors = append(errors, fieldErr.Field()+" must be one of: "+fieldErr.Param())
		default:
			errors = append(errors, fieldErr.Error())
		}
	}
	if len(errors) > 0 {
		return errors, nil
	}
	return nil, nil
}

type EditPostResponse struct {
	PostId string `json:"postId"`
}

func NewEditPostResponse(postId string) *EditPostResponse {
	return &EditPostResponse{
		PostId: postId,
	}
}
