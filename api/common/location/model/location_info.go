package model

import (
	"encoding/json"
	"strings"
)

// sql model for user location info
type UserLocationInfo struct {
	Country   string  `bson:"country_name" json:"country"`
	City      string  `bson:"city_name" json:"city"`
	Lat       float64 `bson:"latitude" json:"latitude"`
	Lon       float64 `bson:"longitude" json:"longitude"`
	Timezone  string  `bson:"timezone" json:"timezone"`
	GmtOffset string  `bson:"gmt_offset" json:"gmt_offset"`
	Locale    string  `bson:"locale" json:"locale"`
}

func NewUserLocationInfo(
	country, city string, lat, lon float64, tz string, gmtOffset string, locale string,
) *UserLocationInfo {
	return &UserLocationInfo{
		Country:   country,
		City:      city,
		Lat:       lat,
		Lon:       lon,
		Timezone:  tz,
		GmtOffset: gmtOffset,
		Locale:    locale,
	}
}

// unmarshal json to UserLocationInfo
func (l *UserLocationInfo) UnmarshalJSON(data []byte) error {
	type Alias UserLocationInfo
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	l.Country = strings.TrimSpace(l.Country)
	l.City = strings.TrimSpace(l.City)
	l.Timezone = strings.TrimSpace(l.Timezone)
	l.GmtOffset = strings.TrimSpace(l.GmtOffset)
	l.Locale = strings.TrimSpace(l.Locale)
	return nil
}

func (l *UserLocationInfo) MarshalJSON() ([]byte, error) {
	type Alias UserLocationInfo
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	return json.Marshal(aux)
}
