package model

// sql model for user location info
type UserLocationInfo struct {
	Country  string `bson:"country_name" json:"country"`
	City     string `bson:"city_name" json:"city"`
	Lat      string `bson:"latitude" json:"latitude"`
	Lon      string `bson:"longitude" json:"longitude"`
	Accuracy string `bson:"accuracy_radius" json:"accuracy_radius"`
}
