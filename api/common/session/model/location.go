package model

type LocationInfo struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	Country   string  `json:"country" bson:"country"`
	City      string  `json:"city" bson:"city"`
}

func NewLocationInfo(latitude, longitude float64, country, city string) *LocationInfo {
	return &LocationInfo{
		Latitude:  latitude,
		Longitude: longitude,
		Country:   country,
		City:      city,
	}
}
