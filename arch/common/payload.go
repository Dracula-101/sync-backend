package common

import (
	"github.com/gin-gonic/gin"
)

const (
	payloadUser       string = "User"
	payloadDeviceId   string = "X-Device-Id"
	payloadDeviceName string = "X-Device-Name"
)

type ContextPayload interface {
	MustGetUserId(ctx *gin.Context) *string
	SetUserId(ctx *gin.Context, value string)

	MustGetIP(ctx *gin.Context) string
	MustGetUserAgent(ctx *gin.Context) string
	MustGetDeviceId(ctx *gin.Context) string
	MustGetDeviceName(ctx *gin.Context) string
}

type payload struct{}

func NewContextPayload() ContextPayload {
	return &payload{}
}

func (payload *payload) SetUserId(ctx *gin.Context, value string) {
	ctx.Set(payloadUser, value)
}

func (payload *payload) MustGetUserId(ctx *gin.Context) *string {
	value, ok := ctx.MustGet(payloadUser).(string)
	if !ok {
		return nil
	}
	return &value
}

func (payload *payload) MustGetIP(ctx *gin.Context) string {
	return ctx.ClientIP()
}

func (payload *payload) MustGetUserAgent(ctx *gin.Context) string {
	return ctx.Request.UserAgent()
}

func (payload *payload) MustGetDeviceId(ctx *gin.Context) string {
	value, ok := ctx.Get(payloadDeviceId)
	if !ok {
		return "default-device-id"
	}
	return value.(string)
}

func (payload *payload) MustGetDeviceName(ctx *gin.Context) string {
	value := ctx.GetHeader(payloadDeviceName)
	if value == "" {
		return "default-device-name"
	}
	return value
}
