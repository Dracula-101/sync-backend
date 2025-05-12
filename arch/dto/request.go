package coredto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type BaseDeviceRequest struct {
	DeviceId        string `json:"device_id"`
	DeviceName      string `json:"device_name"`
	DeviceType      string `json:"device_type"`
	DeviceOS        string `json:"device_os"`
	DeviceModel     string `json:"device_model"`
	DeviceVersion   string `json:"device_version"`
	DeviceUserAgent string `json:"device_user_agent"`
}

func (b *BaseDeviceRequest) GetValue() *BaseDeviceRequest {
	return b
}

func (b *BaseDeviceRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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

type BaseLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Locale    string  `json:"locale"`
	TimeZone  string  `json:"timezone"`
	GMTOffset string  `json:"gmt_offset"`
	IpAddress string  `json:"ip_address"`
}

func (b *BaseLocationRequest) GetValue() *BaseLocationRequest {
	return b
}

func (b *BaseLocationRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
