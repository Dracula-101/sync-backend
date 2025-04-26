package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type LoginHistory struct {
	LoginTime primitive.DateTime `bson:"loginTime" json:"loginTime"`
	IpAddress string             `bson:"ipAddress" json:"ipAddress"`
	Location  UserLocationInfo   `bson:"location" json:"location"`
	UserAgent string             `bson:"userAgent" json:"userAgent"`
	Device    UserDeviceInfo     `bson:"device" json:"device"`
}

type UserLocationInfo struct {
	Country string `bson:"country" json:"country"`
	Region  string `bson:"region" json:"region"`
}

type UserDeviceInfo struct {
	Os    string `bson:"os" json:"os"`
	Type  string `bson:"type" json:"type"`
	Name  string `bson:"name" json:"name"`
	Model string `bson:"model" json:"model"`
}
