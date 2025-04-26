package dto

import (
	coredto "sync-backend/arch/dto"

	"github.com/go-playground/validator/v10"
)

type LocationRequest struct {
	coredto.BaseRequest
}

func NewLocationRequest() *LocationRequest {
	return &LocationRequest{}
}

func (l *LocationRequest) GetValue() *LocationRequest {
	return l
}

func (l *LocationRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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

type LocationResponse struct {
	Country  string `json:"country"`
	Region   string `json:"region"`
	Lat      string `json:"lat"`
	Lon      string `json:"lon"`
	Accuracy int    `json:"accuracy"`
}

func NewLocationResponse(country string, region string, lat string, lon string, accuracy int) *LocationResponse {
	return &LocationResponse{
		Country:  country,
		Region:   region,
		Lat:      lat,
		Lon:      lon,
		Accuracy: accuracy,
	}
}

func (l *LocationResponse) GetValue() *LocationResponse {
	return l
}

func (l *LocationResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
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
