package common

import (
	coredto "sync-backend/arch/dto"
	"sync-backend/arch/network"

	"github.com/gin-gonic/gin"
)

type ContextPayload interface {
	MustGetUserId(ctx *gin.Context) *string
	SetUserId(ctx *gin.Context, value string)

	MustGetSessionId(ctx *gin.Context) string
	SetSessionId(ctx *gin.Context, value string)

	MustGetIP(ctx *gin.Context) string
	MustGetUserAgent(ctx *gin.Context) string
	SetRequestDeviceDetails(ctx *gin.Context, req *coredto.BaseDeviceRequest)
	SetRequestLocationDetails(ctx *gin.Context, req *coredto.BaseLocationRequest)
}

type payload struct{}

func NewContextPayload() ContextPayload {
	return &payload{}
}

func (payload *payload) MustGetUserId(ctx *gin.Context) *string {
	value, ok := ctx.MustGet(network.UserPayload).(string)
	if !ok {
		return nil
	}
	return &value
}

func (payload *payload) SetUserId(ctx *gin.Context, value string) {
	ctx.Set(network.UserPayload, value)
}

func (payload *payload) MustGetSessionId(ctx *gin.Context) string {
	value, ok := ctx.Get(network.SessionIdHeader)
	if !ok {
		return ""
	}
	return value.(string)
}

func (payload *payload) SetSessionId(ctx *gin.Context, value string) {
	ctx.Set(network.SessionIdHeader, value)
}

func (payload *payload) MustGetIP(ctx *gin.Context) string {
	return ctx.ClientIP()
}

func (payload *payload) MustGetUserAgent(ctx *gin.Context) string {
	return ctx.Request.UserAgent()
}

func (payload *payload) MustGetDeviceId(ctx *gin.Context) string {
	value := ctx.GetHeader(network.DeviceIdHeader)
	if value == "" {
		return network.DefaultDeviceId
	}
	return value
}

func (payload *payload) MustGetDeviceName(ctx *gin.Context) string {
	value := ctx.GetHeader(network.DeviceNameHeader)
	if value == "" {
		return network.DefaultDeviceName
	}
	return value
}

func (payload *payload) MustGetDeviceType(ctx *gin.Context) string {
	value := ctx.GetHeader(network.DeviceTypeHeader)
	if value == "" {
		return network.DefaultDeviceType
	}
	return value
}

func (payload *payload) MustGetDeviceOS(ctx *gin.Context) string {
	value := ctx.GetHeader(network.DeviceOsHeader)
	if value == "" {
		return network.DefaultDeviceOs
	}
	return value
}

func (payload *payload) MustGetDeviceModel(ctx *gin.Context) string {
	value := ctx.GetHeader(network.DeviceModelHeader)
	if value == "" {
		return network.DefaultDeviceModel
	}
	return value
}

func (payload *payload) MustGetDeviceVersion(ctx *gin.Context) string {
	value := ctx.GetHeader(network.DeviceVersionHeader)
	if value == "" {
		return network.DefaultDeviceVersion
	}
	return value
}

func (payload *payload) SetRequestDeviceDetails(ctx *gin.Context, req *coredto.BaseDeviceRequest) {
	req.DeviceId = payload.MustGetDeviceId(ctx)
	req.DeviceName = payload.MustGetDeviceName(ctx)
	req.DeviceType = payload.MustGetDeviceType(ctx)
	req.DeviceOS = payload.MustGetDeviceOS(ctx)
	req.DeviceModel = payload.MustGetDeviceModel(ctx)
	req.DeviceVersion = payload.MustGetDeviceVersion(ctx)
	req.DeviceUserAgent = payload.MustGetUserAgent(ctx)
}

func (payload *payload) SetRequestLocationDetails(ctx *gin.Context, req *coredto.BaseLocationRequest) {
	locationData, exists := ctx.Get(network.UserLocation)
	if !exists {
		return
	}
	if locationData == nil {
		return
	}
	location := locationData.(*coredto.BaseLocationRequest)
	req.Country = location.Country
	req.City = location.City
	req.Latitude = location.Latitude
	req.Longitude = location.Longitude
	req.TimeZone = location.TimeZone
	req.GMTOffset = location.GMTOffset
	req.Locale = location.Locale
	req.IpAddress = location.IpAddress
}
