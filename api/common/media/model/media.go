package model

type MediaInfo struct {
	Id     string `json:"id" bson:"id"`
	Type   string `json:"type" bson:"type"`
	Url    string `json:"url" bson:"url"`
	Width  int    `json:"width" bson:"width"`
	Height int    `json:"height" bson:"height"`
	Size   int64  `json:"size" bson:"size"`
}

type MediaMimeType string

const (
	MediaMimeTypeImage    MediaMimeType = "image"
	MediaMimeTypeVideo    MediaMimeType = "video"
	MediaMimeTypeAudio    MediaMimeType = "audio"
	MediaMimeTypeDocument MediaMimeType = "raw"
)
