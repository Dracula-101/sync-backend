package model

import (
	"sync-backend/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const EngagementCollectionName = "engagements"
const PostAnalyticsCollectionName = "post_analytics"

// EngagementType defines the type of interaction a user has with a post
type EngagementType string

const (
	EngagementTypeView     EngagementType = "view"
	EngagementTypeUpvote   EngagementType = "upvote"
	EngagementTypeDownvote EngagementType = "downvote"
	EngagementTypeShare    EngagementType = "share"
	EngagementTypeSave     EngagementType = "save"
	EngagementTypeClick    EngagementType = "click"
	EngagementTypeComment  EngagementType = "comment"
	EngagementTypeAward    EngagementType = "award"
	EngagementTypeReport   EngagementType = "report"
)

// EngagementSource defines where the engagement occurred
type EngagementSource string

const (
	SourceFeed         EngagementSource = "feed"
	SourceProfile      EngagementSource = "profile"
	SourceCommunity    EngagementSource = "community"
	SourceSearch       EngagementSource = "search"
	SourceNotification EngagementSource = "notification"
	SourceDirectLink   EngagementSource = "direct_link"
	SourceExternal     EngagementSource = "external"
)

// Engagement represents a single interaction between a user and a post/comment
type Engagement struct {
	Id           primitive.ObjectID     `bson:"_id,omitempty" json:"-"`
	EngagementId string                 `bson:"engagementId" json:"id"`
	UserId       string                 `bson:"userId" json:"userId"`
	PostId       string                 `bson:"postId,omitempty" json:"postId,omitempty"`
	CommentId    string                 `bson:"commentId,omitempty" json:"commentId,omitempty"`
	Type         EngagementType         `bson:"type" json:"type"`
	Source       EngagementSource       `bson:"source" json:"source"`
	DeviceInfo   DeviceInfo             `bson:"deviceInfo" json:"deviceInfo"`
	LocationInfo *LocationInfo          `bson:"locationInfo,omitempty" json:"locationInfo,omitempty"`
	Duration     int                    `bson:"duration,omitempty" json:"duration,omitempty"`
	Metadata     map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt    primitive.DateTime     `bson:"createdAt" json:"createdAt"`
}

// DeviceInfo represents information about the user's device
type DeviceInfo struct {
	Type       string `bson:"type" json:"type"`
	OS         string `bson:"os" json:"os"`
	Browser    string `bson:"browser" json:"browser"`
	AppVersion string `bson:"appVersion,omitempty" json:"appVersion,omitempty"`
	IpAddress  string `bson:"ipAddress" json:"ipAddress"`
	UserAgent  string `bson:"userAgent" json:"userAgent"`
}

// LocationInfo represents the user's location information
type LocationInfo struct {
	Country   string  `bson:"country,omitempty" json:"country,omitempty"`
	City      string  `bson:"city,omitempty" json:"city,omitempty"`
	Region    string  `bson:"region,omitempty" json:"region,omitempty"`
	Longitude float64 `bson:"longitude,omitempty" json:"longitude,omitempty"`
	Latitude  float64 `bson:"latitude,omitempty" json:"latitude,omitempty"`
	TimeZone  string  `bson:"timeZone,omitempty" json:"timeZone,omitempty"`
}

// PostAnalytics represents aggregated analytics for a post
type PostAnalytics struct {
	Id                primitive.ObjectID       `bson:"_id,omitempty" json:"-"`
	AnalyticsId       string                   `bson:"analyticsId" json:"id"`
	PostId            string                   `bson:"postId" json:"postId"`
	ViewCount         int                      `bson:"viewCount" json:"viewCount"`
	UniqueViewerCount int                      `bson:"uniqueViewerCount" json:"uniqueViewerCount"`
	AvgViewDuration   float64                  `bson:"avgViewDuration" json:"avgViewDuration"`
	EngagementRate    float64                  `bson:"engagementRate" json:"engagementRate"`
	ConversionRate    float64                  `bson:"conversionRate" json:"conversionRate"`
	ViewsByDevice     map[string]int           `bson:"viewsByDevice" json:"viewsByDevice"`
	ViewsByCountry    map[string]int           `bson:"viewsByCountry" json:"viewsByCountry"`
	ViewsBySource     map[EngagementSource]int `bson:"viewsBySource" json:"viewsBySource"`
	EngagementsByType map[EngagementType]int   `bson:"engagementsByType" json:"engagementsByType"`
	ViewTrend         map[string]int           `bson:"viewTrend" json:"viewTrend"`
	SynergyTrend      map[string]int           `bson:"synergyTrend" json:"synergyTrend"`
	UpdatedAt         primitive.DateTime       `bson:"updatedAt" json:"updatedAt"`
}

// NewEngagement creates a new engagement record
func NewEngagement(userId string, type_ EngagementType, source EngagementSource, deviceInfo DeviceInfo) *Engagement {
	now := primitive.NewDateTimeFromTime(time.Now())

	return &Engagement{
		Id:           primitive.NewObjectID(),
		EngagementId: utils.GenerateUUID(),
		UserId:       userId,
		Type:         type_,
		Source:       source,
		DeviceInfo:   deviceInfo,
		CreatedAt:    now,
	}
}

// NewPostAnalytics creates a new post analytics record
func NewPostAnalytics(postId string) *PostAnalytics {
	now := primitive.NewDateTimeFromTime(time.Now())

	return &PostAnalytics{
		Id:                primitive.NewObjectID(),
		AnalyticsId:       utils.GenerateUUID(),
		PostId:            postId,
		ViewCount:         0,
		UniqueViewerCount: 0,
		AvgViewDuration:   0,
		EngagementRate:    0,
		ConversionRate:    0,
		ViewsByDevice:     make(map[string]int),
		ViewsByCountry:    make(map[string]int),
		ViewsBySource:     make(map[EngagementSource]int),
		EngagementsByType: make(map[EngagementType]int),
		ViewTrend:         make(map[string]int),
		SynergyTrend:      make(map[string]int),
		UpdatedAt:         now,
	}
}
