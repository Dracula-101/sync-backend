package model

type MediaInfo struct {
	Id     string `json:"id" bson:"id"`
	Url    string `json:"url" bson:"url"`
	Width  int    `json:"width" bson:"width"`
	Height int    `json:"height" bson:"height"`
}

type MediaMimeType string

const (
	MediaMimeTypeImage    MediaMimeType = "image"
	MediaMimeTypeVideo    MediaMimeType = "video"
	MediaMimeTypeAudio    MediaMimeType = "audio"
	MediaMimeTypeDocument MediaMimeType = "raw"
)
