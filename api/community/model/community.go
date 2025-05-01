package model

import (
	"math/rand"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Community struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Slug        string             `bson:"slug" json:"slug"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	ShortDesc   string             `bson:"shortDesc" json:"shortDesc"`
	OwnerId     string             `bson:"ownerId" json:"ownerId"`
	IsPrivate   bool               `bson:"isPrivate" json:"isPrivate"`
	MemberCount int64              `bson:"memberCount" json:"memberCount"`
	PostCount   int64              `bson:"postCount" json:"postCount"`
	Media       CommunityMedia     `bson:"media" json:"media"`
	Rules       []CommunityRule    `bson:"rules" json:"rules"`
	Tags        []CommunityTag     `bson:"tags" json:"tags"`
	Moderators  []string           `bson:"moderators" json:"moderators"`
	Settings    CommunitySettings  `bson:"settings" json:"settings"`
	Stats       CommunityStats     `bson:"stats" json:"stats"`
	Status      string             `bson:"status" json:"status"`
	Metadata    Metadata           `bson:"metadata" json:"metadata"`
}

type CommunityMedia struct {
	Avatar       Image   `bson:"avatar" json:"avatar"`
	Background   Image   `bson:"background" json:"background"`
	Favicon      string  `bson:"favicon" json:"favicon"`
	Gallery      []Image `bson:"gallery" json:"gallery"`
	FeaturedPost string  `bson:"featuredPost" json:"featuredPost"`
}

type Image struct {
	ID           string `bson:"id" json:"id"`
	Url          string `bson:"url" json:"url"`
	Width        int    `bson:"width" json:"width"`
	Height       int    `bson:"height" json:"height"`
	ThumbnailUrl string `bson:"thumbnailUrl,omitempty" json:"thumbnailUrl,omitempty"`
	AltText      string `bson:"altText" json:"altText"`
	Caption      string `bson:"caption,omitempty" json:"caption,omitempty"`
	FileSize     int64  `bson:"fileSize,omitempty" json:"fileSize,omitempty"`
	MimeType     string `bson:"mimeType,omitempty" json:"mimeType,omitempty"`
}

type CommunityRule struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title        string             `bson:"title" json:"title"`
	Description  string             `bson:"description" json:"description"`
	Priority     int                `bson:"priority" json:"priority"`
	IsRequired   bool               `bson:"isRequired" json:"isRequired"`
	ReportOption bool               `bson:"reportOption" json:"reportOption"`
	Metadata     Metadata           `bson:"metadata" json:"metadata"`
}

type CommunityTag struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Icon        string             `bson:"icon" json:"icon"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Color       string             `bson:"color" json:"color"`
	PostCount   int64              `bson:"postCount" json:"postCount"`
	IsOfficial  bool               `bson:"isOfficial" json:"isOfficial"`
	Metadata    Metadata           `bson:"metadata" json:"metadata"`
}

type CommunitySettings struct {
	JoinPolicy           string   `bson:"joinPolicy" json:"joinPolicy"`
	PostApproval         bool     `bson:"postApproval" json:"postApproval"`
	AllowedPostTypes     []string `bson:"allowedPostTypes" json:"allowedPostTypes"`
	EnableDirectMessages bool     `bson:"enableDirectMessages" json:"enableDirectMessages"`
	ShowInDiscovery      bool     `bson:"showInDiscovery" json:"showInDiscovery"`
	EnableComments       bool     `bson:"enableComments" json:"enableComments"`
	DefaultCommentSort   string   `bson:"defaultCommentSort" json:"defaultCommentSort"`
	DefaultPostSort      string   `bson:"defaultPostSort" json:"defaultPostSort"`
	EnablePolls          bool     `bson:"enablePolls" json:"enablePolls"`
	EnableEvents         bool     `bson:"enableEvents" json:"enableEvents"`
	AllowCrossposting    bool     `bson:"allowCrossposting" json:"allowCrossposting"`
	AllowUserFlairs      bool     `bson:"allowUserFlairs" json:"allowUserFlairs"`
	AllowPostFlairs      bool     `bson:"allowPostFlairs" json:"allowPostFlairs"`
	AllowNSFWContent     bool     `bson:"allowNSFWContent" json:"allowNSFWContent"`
	RequirePostTag       bool     `bson:"requirePostTag" json:"requirePostTag"`
	Language             string   `bson:"language" json:"language"`
	ContentFilters       []string `bson:"contentFilters" json:"contentFilters"`
	MinAccountAgeToPost  int      `bson:"minAccountAgeToPost" json:"minAccountAgeToPost"`
	MinKarmaToPost       int      `bson:"minKarmaToPost" json:"minKarmaToPost"`
}

type CommunityStats struct {
	DailyActiveUsers   int64              `bson:"dailyActiveUsers" json:"dailyActiveUsers"`
	WeeklyActiveUsers  int64              `bson:"weeklyActiveUsers" json:"weeklyActiveUsers"`
	MonthlyActiveUsers int64              `bson:"monthlyActiveUsers" json:"monthlyActiveUsers"`
	GrowthRate         float64            `bson:"growthRate" json:"growthRate"`
	EngagementRate     float64            `bson:"engagementRate" json:"engagementRate"`
	TopTags            []TagStats         `bson:"topTags" json:"topTags"`
	PopularityScore    float64            `bson:"popularityScore" json:"popularityScore"`
	LastUpdated        primitive.DateTime `bson:"lastUpdated" json:"lastUpdated"`
}

type TagStats struct {
	TagID      string  `bson:"tagId" json:"tagId"`
	Name       string  `bson:"name" json:"name"`
	PostCount  int64   `bson:"postCount" json:"postCount"`
	Popularity float64 `bson:"popularity" json:"popularity"`
}

type CommunityMember struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	CommunityID   string              `bson:"communityId" json:"communityId"`
	UserID        string              `bson:"userId" json:"userId"`
	Role          string              `bson:"role" json:"role"`
	JoinDate      primitive.DateTime  `bson:"joinDate" json:"joinDate"`
	Status        string              `bson:"status" json:"status"`
	DisplayName   string              `bson:"displayName" json:"displayName"`
	Flair         MemberFlair         `bson:"flair" json:"flair"`
	Contributions MemberContributions `bson:"contributions" json:"contributions"`
	LastActive    primitive.DateTime  `bson:"lastActive" json:"lastActive"`
	Metadata      Metadata            `bson:"metadata" json:"metadata"`
}

type MemberFlair struct {
	Text  string `bson:"text" json:"text"`
	Color string `bson:"color" json:"color"`
	Icon  string `bson:"icon" json:"icon"`
}

type MemberContributions struct {
	PostCount         int64   `bson:"postCount" json:"postCount"`
	CommentCount      int64   `bson:"commentCount" json:"commentCount"`
	ReactionCount     int64   `bson:"reactionCount" json:"reactionCount"`
	KarmaPoints       int64   `bson:"karmaPoints" json:"karmaPoints"`
	ContributionScore float64 `bson:"contributionScore" json:"contributionScore"`
}

type CommunityInvite struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CommunityID string             `bson:"communityId" json:"communityId"`
	InviterID   string             `bson:"inviterId" json:"inviterId"`
	InviteeID   string             `bson:"inviteeId,omitempty" json:"inviteeId,omitempty"`
	Email       string             `bson:"email,omitempty" json:"email,omitempty"`
	Code        string             `bson:"code" json:"code"`
	Status      string             `bson:"status" json:"status"`
	Role        string             `bson:"role" json:"role"`
	ExpiresAt   primitive.DateTime `bson:"expiresAt" json:"expiresAt"`
	Metadata    Metadata           `bson:"metadata" json:"metadata"`
}

type CommunityJoinRequest struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CommunityID string             `bson:"communityId" json:"communityId"`
	UserID      string             `bson:"userId" json:"userId"`
	Message     string             `bson:"message" json:"message"`
	Status      string             `bson:"status" json:"status"`
	ReviewedBy  string             `bson:"reviewedBy,omitempty" json:"reviewedBy,omitempty"`
	Metadata    Metadata           `bson:"metadata" json:"metadata"`
}

type CommunityBan struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CommunityID string             `bson:"communityId" json:"communityId"`
	UserID      string             `bson:"userId" json:"userId"`
	ModeratorID string             `bson:"moderatorId" json:"moderatorId"`
	Reason      string             `bson:"reason" json:"reason"`
	Duration    int                `bson:"duration,omitempty" json:"duration,omitempty"`
	ExpiresAt   primitive.DateTime `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
	IsActive    bool               `bson:"isActive" json:"isActive"`
	Metadata    Metadata           `bson:"metadata" json:"metadata"`
}

type ModAction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CommunityID string             `bson:"communityId" json:"communityId"`
	ModeratorID string             `bson:"moderatorId" json:"moderatorId"`
	ActionType  string             `bson:"actionType" json:"actionType"`
	Details     string             `bson:"details" json:"details"`
	Notes       string             `bson:"notes,omitempty" json:"notes,omitempty"`
	Metadata    Metadata           `bson:"metadata" json:"metadata"`
}

type CommunityEvent struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CommunityID    string             `bson:"communityId" json:"communityId"`
	CreatorID      string             `bson:"creatorId" json:"creatorId"`
	Title          string             `bson:"title" json:"title"`
	Description    string             `bson:"description" json:"description"`
	StartTime      primitive.DateTime `bson:"startTime" json:"startTime"`
	EndTime        primitive.DateTime `bson:"endTime" json:"endTime"`
	Location       EventLocation      `bson:"location" json:"location"`
	IsVirtual      bool               `bson:"isVirtual" json:"isVirtual"`
	MeetingURL     string             `bson:"meetingUrl,omitempty" json:"meetingUrl,omitempty"`
	Cover          Image              `bson:"cover,omitempty" json:"cover,omitempty"`
	MaxAttendees   int                `bson:"maxAttendees,omitempty" json:"maxAttendees,omitempty"`
	AttendeesCount int                `bson:"attendeesCount" json:"attendeesCount"`
	Status         string             `bson:"status" json:"status"`
	Metadata       Metadata           `bson:"metadata" json:"metadata"`
}

type EventLocation struct {
	Address    string  `bson:"address,omitempty" json:"address,omitempty"`
	City       string  `bson:"city,omitempty" json:"city,omitempty"`
	State      string  `bson:"state,omitempty" json:"state,omitempty"`
	Country    string  `bson:"country,omitempty" json:"country,omitempty"`
	PostalCode string  `bson:"postalCode,omitempty" json:"postalCode,omitempty"`
	Latitude   float64 `bson:"latitude,omitempty" json:"latitude,omitempty"`
	Longitude  float64 `bson:"longitude,omitempty" json:"longitude,omitempty"`
	VenueName  string  `bson:"venueName,omitempty" json:"venueName,omitempty"`
}

type Metadata struct {
	CreatedAt  primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt  primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
	DeletedAt  primitive.DateTime `bson:"deletedAt,omitempty" json:"-"`
	CreatedBy  string             `bson:"createdBy" json:"createdBy"`
	UpdatedBy  string             `bson:"updatedBy,omitempty" json:"updatedBy,omitempty"`
	DeletedBy  string             `bson:"deletedBy,omitempty" json:"-"`
	IPAddress  string             `bson:"ipAddress,omitempty" json:"-"`
	UserAgent  string             `bson:"userAgent,omitempty" json:"-"`
	Version    int                `bson:"version" json:"version"`
	CustomData map[string]any     `bson:"customData,omitempty" json:"customData,omitempty"`
}

func NewCommunity(name, description, ownerId string, avatarUrl *string, backgroundUrl *string, tags []CommunityTag) *Community {
	now := primitive.NewDateTimeFromTime(time.Now())
	slug := generateSlug(name)
	communityAvatarUrl := getDefaultAvatarUrl()
	if avatarUrl != nil {
		communityAvatarUrl = *avatarUrl
	}
	communityBackgroundUrl := getDefaultBackgroundUrl()
	if backgroundUrl != nil {
		communityBackgroundUrl = *backgroundUrl
	}

	return &Community{
		ID:          primitive.NewObjectID(),
		Slug:        slug,
		Name:        name,
		Description: description,
		ShortDesc:   truncateString(description, 160),
		OwnerId:     ownerId,
		IsPrivate:   false,
		MemberCount: 1,
		PostCount:   0,
		Moderators:  []string{ownerId},
		Media: CommunityMedia{
			Avatar: Image{
				ID:       primitive.NewObjectID().Hex(),
				Url:      communityAvatarUrl,
				Width:    512,
				Height:   512,
				AltText:  name + " community avatar",
				MimeType: "image/png",
			},
			Background: Image{
				ID:       primitive.NewObjectID().Hex(),
				Url:      communityBackgroundUrl,
				Width:    1920,
				Height:   1080,
				AltText:  name + " community background",
				MimeType: "image/jpeg",
			},
			Gallery:      []Image{},
			FeaturedPost: "",
		},
		Rules: []CommunityRule{
			{
				ID:           primitive.NewObjectID(),
				Title:        "Be respectful",
				Description:  "Treat others with respect. No personal attacks, harassment, or hate speech.",
				Priority:     1,
				IsRequired:   true,
				ReportOption: true,
				Metadata: Metadata{
					CreatedAt: now,
					UpdatedAt: now,
					CreatedBy: ownerId,
					Version:   1,
				},
			},
		},
		Tags:   tags,
		Status: "active",
		Settings: CommunitySettings{
			JoinPolicy:           "open",
			PostApproval:         false,
			AllowedPostTypes:     []string{"text", "image", "link", "poll"},
			EnableDirectMessages: true,
			ShowInDiscovery:      true,
			EnableComments:       true,
			DefaultCommentSort:   "top",
			DefaultPostSort:      "hot",
			EnablePolls:          true,
			EnableEvents:         true,
			AllowCrossposting:    true,
			AllowUserFlairs:      true,
			AllowPostFlairs:      true,
			AllowNSFWContent:     false,
			RequirePostTag:       false,
			Language:             "en",
			ContentFilters:       []string{},
			MinAccountAgeToPost:  0,
			MinKarmaToPost:       0,
		},
		Stats: CommunityStats{
			DailyActiveUsers:   1,
			WeeklyActiveUsers:  1,
			MonthlyActiveUsers: 1,
			GrowthRate:         0.0,
			EngagementRate:     0.0,
			TopTags:            []TagStats{},
			PopularityScore:    0.0,
			LastUpdated:        now,
		},
		Metadata: Metadata{
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: ownerId,
			Version:   1,
		},
	}
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")

	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")

	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "community-" + randomString(8)
	}

	return slug
}

func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	lastSpace := strings.LastIndex(s[:maxLength], " ")
	if lastSpace == -1 {
		return s[:maxLength] + "..."
	}

	return s[:lastSpace] + "..."
}

func getDefaultAvatarUrl() string {
	return "https://www.shutterstock.com/image-vector/sound-wave-simple-icon-black-260nw-2126255804.jpg"
}

func getDefaultBackgroundUrl() string {
	return "https://placehold.co/1200x400.png"
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
