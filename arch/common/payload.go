package common

import (
	"github.com/gin-gonic/gin"

	userModel "sync-backend/api/user/model"
)

const (
	payloadUser = "user"
)

type ContextPayload interface {
	MustGetUserId(ctx *gin.Context) *userModel.User
	SetUserId(ctx *gin.Context, value *userModel.User)
}

type payload struct{}

func NewContextPayload() ContextPayload {
	return &payload{}
}

func (payload *payload) SetUserId(ctx *gin.Context, value *userModel.User) {
	ctx.Set(payloadUser, value.ID)
}

func (payload *payload) MustGetUserId(ctx *gin.Context) *userModel.User {
	value, ok := ctx.MustGet(payloadUser).(*userModel.User)
	if !ok {
		return nil
	}
	return value
}
