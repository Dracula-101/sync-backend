package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type DeviceToken struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DeviceId string             `bson:"deviceId" json:"deviceId"`
	Token    string             `bson:"token" json:"token"`
	Type     string             `bson:"type" json:"type"`
}

func NewDeviceToken(token string, deviceId string, tokenType string) *DeviceToken {
	return &DeviceToken{
		Token:    token,
		DeviceId: deviceId,
		Type:     tokenType,
	}
}
