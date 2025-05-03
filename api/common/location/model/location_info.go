package model

// sql model for user location info
type UserLocationInfo struct {
	Country  string  `bson:"country_name" json:"country"`
	City     string  `bson:"city_name" json:"city"`
	Lat      float64 `bson:"latitude" json:"latitude"`
	Lon      float64 `bson:"longitude" json:"longitude"`
	Accuracy string  `bson:"accuracy_radius" json:"accuracy_radius"`
}
