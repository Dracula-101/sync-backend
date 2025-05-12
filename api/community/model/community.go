package model

import (
	"context"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"sync-backend/arch/mongo"
)

const CommunityCollectionName = "communities"

type Community struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	CommunityId string             `bson:"communityId" json:"id"`
	Slug        string             `bson:"slug" json:"slug"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	ShortDesc   string             `bson:"shortDesc" json:"shortDesc"`
	OwnerId     string             `bson:"ownerId" json:"ownerId"`
	IsPrivate   bool               `bson:"isPrivate" json:"isPrivate"`
	Members     []string           `bson:"members" json:"members"`
	MemberCount int64              `bson:"memberCount" json:"memberCount"`
	PostCount   int64              `bson:"postCount" json:"postCount"`
	Media       CommunityMedia     `bson:"media" json:"media"`
	Rules       []CommunityRule    `bson:"rules" json:"rules"`
	Tags        []CommunityTagInfo `bson:"tags" json:"tags"`
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
	ID     string `bson:"id" json:"id"`
	Url    string `bson:"url" json:"url"`
	Width  int    `bson:"width" json:"width"`
	Height int    `bson:"height" json:"height"`
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
	MinSynergyToPost     int      `bson:"minSynergyToPost" json:"minSynergyToPost"`
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
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"-"`
	CommunityID   string              `bson:"communityId" json:"id"`
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
	SynergyPoints     int64   `bson:"synergyPoints" json:"synergyPoints"`
	ContributionScore float64 `bson:"contributionScore" json:"contributionScore"`
}

type CommunityInvite struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"-"`
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
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	CommunityID string             `bson:"communityId" json:"communityId"`
	UserID      string             `bson:"userId" json:"userId"`
	Message     string             `bson:"message" json:"message"`
	Status      string             `bson:"status" json:"status"`
	ReviewedBy  string             `bson:"reviewedBy,omitempty" json:"reviewedBy,omitempty"`
	Metadata    Metadata           `bson:"metadata" json:"metadata"`
}

type CommunityBan struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"-"`
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
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	CommunityID string             `bson:"communityId" json:"communityId"`
	ModeratorID string             `bson:"moderatorId" json:"moderatorId"`
	ActionType  string             `bson:"actionType" json:"actionType"`
	Details     string             `bson:"details" json:"details"`
	Notes       string             `bson:"notes,omitempty" json:"notes,omitempty"`
	Metadata    Metadata           `bson:"metadata" json:"metadata"`
}

type CommunityEvent struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"-"`
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

type NewCommunityArgs struct {
	Name        string
	Description string
	OwnerId     string
	Avatar      Image
	Background  Image
	Tags        []CommunityTagInfo
}

func NewCommunity(args NewCommunityArgs) *Community {
	now := primitive.NewDateTimeFromTime(time.Now())
	slug := generateSlug(args.Name)

	return &Community{
		ID:          primitive.NewObjectID(),
		CommunityId: uuid.New().String(),
		Slug:        slug,
		Name:        args.Name,
		Description: args.Description,
		ShortDesc:   truncateString(args.Description, 160),
		OwnerId:     args.OwnerId,
		IsPrivate:   false,
		Members:     []string{args.OwnerId},
		MemberCount: 1,
		PostCount:   0,
		Moderators:  []string{args.OwnerId},
		Media: CommunityMedia{
			Avatar:       args.Avatar,
			Background:   args.Background,
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
					CreatedBy: args.OwnerId,
					Version:   1,
				},
			},
		},
		Tags:   args.Tags,
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
			MinSynergyToPost:     0,
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
			CreatedBy: args.OwnerId,
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

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func (c *Community) GetCollectionName() string {
	return CommunityCollectionName
}

func (c *Community) GetValue() *Community {
	return c
}

func (c *Community) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func (c *Community) EnsureIndexes(db mongo.Database) {
	indexes := []mongod.IndexModel{
		{
			Keys: bson.D{
				{Key: "communityId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_community_id_unique"),
		},
		{
			Keys: bson.D{
				{Key: "slug", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_community_slug_unique"),
		},
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_community_name_unique"),
		},
		{
			Keys: bson.D{
				{Key: "members", Value: 1},
			},
			Options: options.Index().SetName("idx_community_members"),
		},
		{
			Keys: bson.D{
				{Key: "settings.showInDiscovery", Value: 1},
				{Key: "isPrivate", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_community_discovery"),
		},
		{
			Keys: bson.D{
				{Key: "stats.popularityScore", Value: -1},
			},
			Options: options.Index().SetName("idx_community_popularity"),
		},
		// TTL index for deleted communities - 30 days (1 month)
		{
			Keys: bson.D{
				{Key: "metadata.deletedAt", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(30 * 24 * 60 * 60).SetName("ttl_community_deleted"),
		},
	}

	mongo.NewQueryBuilder[Community](db, CommunityCollectionName).Query(context.Background()).CreateIndexes(indexes)
}
