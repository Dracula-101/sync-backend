package model

type DeviceInfo struct {
	DeviceId      string `json:"id" bson:"id"`
	DeviceName    string `json:"name" bson:"name"`
	DeviceType    string `json:"type" bson:"type"`
	DeviceOS      string `json:"os" bson:"os"`
	DeviceModel   string `json:"model" bson:"model"`
	DeviceVersion string `json:"version" bson:"version"`
}

func NewDeviceInfo(deviceId, deviceName, deviceType, deviceOS, deviceModel, deviceVersion string) *DeviceInfo {
	return &DeviceInfo{
		DeviceId:      deviceId,
		DeviceName:    deviceName,
		DeviceType:    deviceType,
		DeviceOS:      deviceOS,
		DeviceModel:   deviceModel,
		DeviceVersion: deviceVersion,
	}
}
