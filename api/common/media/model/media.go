package model

type MediaInfo struct {
	Id           string        `json:"id" bson:"id"`
	Url          string        `json:"url" bson:"url"`
	Size         int           `json:"size" bson:"size"`
	MimeType     MediaMimeType `json:"mimeType" bson:"mimeType"`
	Width        int           `json:"width" bson:"width"`
	Height       int           `json:"height" bson:"height"`
	ThumbnailUrl string        `json:"thumbnailUrl,omitempty" bson:"thumbnailUrl,omitempty"`
}

type MediaMimeType string

const (
	MediaMimeTypeImage    MediaMimeType = "image"
	MediaMimeTypeVideo    MediaMimeType = "video"
	MediaMimeTypeAudio    MediaMimeType = "audio"
	MediaMimeTypeDocument MediaMimeType = "raw"
)
