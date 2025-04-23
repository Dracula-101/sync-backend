package coredto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type BaseRequest struct {
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	DeviceId   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

func (b *BaseRequest) GetValue() *BaseRequest {
	return b
}

func (b *BaseRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
