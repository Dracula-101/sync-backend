package common

import (
	"github.com/gin-gonic/gin"

	userModel "sync-backend/api/user/model"
)

const (
	payloadUser string = "user"
)

type ContextPayload interface {
	MustGetUser(ctx *gin.Context) *userModel.User
	SetUser(ctx *gin.Context, value *userModel.User)
}

type payload struct{}

func NewContextPayload() ContextPayload {
	return &payload{}
}

func (payload *payload) SetUser(ctx *gin.Context, value *userModel.User) {
	ctx.Set(payloadUser, value)
}

func (payload *payload) MustGetUser(ctx *gin.Context) *userModel.User {
	value, ok := ctx.MustGet(payloadUser).(*userModel.User)
	if !ok {
		return nil
	}
	return value
}
