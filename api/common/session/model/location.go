package model

type LocationInfo struct {
	Latitude   float64 `json:"latitude" bson:"latitude"`
	Longitude  float64 `json:"longitude" bson:"longitude"`
	Country    string  `json:"country" bson:"country"`
	City       string  `json:"city" bson:"city"`
	LocaleCode string  `json:"localeCode" bson:"localeCode"`
	Timezone   string  `json:"timezone" bson:"timezone"`
	GmtOffset  string  `json:"gmtOffset" bson:"gmtOffset"`
	IpAddress  string  `json:"ipAddress" bson:"ipAddress"`
}

func NewLocationInfo(latitude, longitude float64, country, city, localeCode, timezone, gmtOffset string, ipAddress string) *LocationInfo {
	return &LocationInfo{
		Latitude:   latitude,
		Longitude:  longitude,
		Country:    country,
		City:       city,
		LocaleCode: localeCode,
		Timezone:   timezone,
		GmtOffset:  gmtOffset,
		IpAddress:  ipAddress,
	}
}
