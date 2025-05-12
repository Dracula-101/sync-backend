package model

type DeviceInfo struct {
	DeviceId        string `json:"id" bson:"id"`
	DeviceName      string `json:"name" bson:"name"`
	DeviceType      string `json:"type" bson:"type"`
	DeviceOS        string `json:"os" bson:"os"`
	DeviceModel     string `json:"model" bson:"model"`
	DeviceVersion   string `json:"version" bson:"version"`
	DeviceUserAgent string `json:"user_agent" bson:"user_agent"`
}
