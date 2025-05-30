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
	MemberCount int64              `bson:"memberCount" json:"memberCount"`
	PostCount   int64              `bson:"postCount" json:"postCount"`
	Media       CommunityMedia     `bson:"media" json:"media"`
	Rules       []CommunityRule    `bson:"rules" json:"rules"`
	Tags        []CommunityTagInfo `bson:"tags" json:"tags"`
	Moderators  []ModeratorInfo    `bson:"moderators" json:"moderators"`
	Settings    CommunitySettings  `bson:"settings" json:"settings"`
	Status      CommunityStatus    `bson:"status" json:"status"`
	Analytics   CommunityAnalytics `bson:"analytics" json:"analytics"`
	CreatedAt   primitive.DateTime `bson:"createdAt" json:"createdAt"`
	UpdatedAt   primitive.DateTime `bson:"updatedAt" json:"updatedAt"`
	Version     int                `bson:"version" json:"version"`
}

type CommunityStatus string

const (
	CommunityStatusActive   CommunityStatus = "active"
	CommunityStatusInactive CommunityStatus = "inactive"
	CommunityStatusDeleted  CommunityStatus = "deleted"
	CommunityStatusBanned   CommunityStatus = "banned"
)

type ModeratorInfo struct {
	UserId  string             `bson:"userId" json:"userId"`
	AddedBy string             `bson:"addedBy" json:"addedBy"`
	AddedAt primitive.DateTime `bson:"addedAt" json:"addedAt"`
	Role    string             `bson:"role" json:"role"`
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
	CreatedAt    primitive.DateTime `bson:"createdAt" json:"createdAt"`
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
	MaxPostsPerDay       int      `bson:"maxPostsPerDay" json:"maxPostsPerDay"`
	AutoModeration       bool     `bson:"autoModeration" json:"autoModeration"`
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
		MemberCount: 1,
		PostCount:   0,
		Moderators:  []ModeratorInfo{},
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
				CreatedAt:    now,
			},
		},
		Tags:      args.Tags,
		Status:    CommunityStatusActive,
		Analytics: *NewCommunityAnalytics(),
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
			MaxPostsPerDay:       50,
			AutoModeration:       false,
		},
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
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
				{Key: "ownerId", Value: 1},
			},
			Options: options.Index().SetName("idx_community_owner"),
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
				{Key: "memberCount", Value: -1},
			},
			Options: options.Index().SetName("idx_community_member_count"),
		},
		{
			Keys: bson.D{
				{Key: "postCount", Value: -1},
			},
			Options: options.Index().SetName("idx_community_post_count"),
		},
		{
			Keys: bson.D{
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("idx_community_created"),
		},
		{
			Keys: bson.D{
				{Key: "tags.name", Value: 1},
			},
			Options: options.Index().SetName("idx_community_tags"),
		},
		// New indexes for analytics
		{
			Keys: bson.D{
				{Key: "analytics.trendingScore", Value: -1},
			},
			Options: options.Index().SetName("idx_community_trending_score"),
		},
		{
			Keys: bson.D{
				{Key: "analytics.engagementScore", Value: -1},
			},
			Options: options.Index().SetName("idx_community_engagement_score"),
		},
		{
			Keys: bson.D{
				{Key: "analytics.qualityScore", Value: -1},
			},
			Options: options.Index().SetName("idx_community_quality_score"),
		},
		{
			Keys: bson.D{
				{Key: "analytics.lastActivityAt", Value: -1},
			},
			Options: options.Index().SetName("idx_community_last_activity"),
		},
		{
			Keys: bson.D{
				{Key: "analytics.activeMembersWeek", Value: -1},
			},
			Options: options.Index().SetName("idx_community_active_members"),
		},
	}
	mongo.NewQueryBuilder[Community](db, CommunityCollectionName).Query(context.Background()).CheckIndexes(indexes)

	searchIndexes := []mongod.SearchIndexModel{
		{
			Definition: bson.D{
				{Key: "mappings", Value: bson.D{
					{Key: "dynamic", Value: false},
					{Key: "fields", Value: bson.D{
						{Key: "name", Value: bson.D{
							{Key: "type", Value: "string"},
							{Key: "analyzer", Value: "standard_lowercase"},
							{Key: "searchAnalyzer", Value: "standard_lowercase"},
						}},
						{Key: "slug", Value: bson.D{
							{Key: "type", Value: "string"},
							{Key: "analyzer", Value: "standard_lowercase"},
							{Key: "searchAnalyzer", Value: "standard_lowercase"},
						}},
						{Key: "shortDesc", Value: bson.D{
							{Key: "type", Value: "string"},
							{Key: "analyzer", Value: "standard_lowercase"},
							{Key: "searchAnalyzer", Value: "standard_lowercase"},
						}},
						{Key: "description", Value: bson.D{
							{Key: "type", Value: "string"},
							{Key: "analyzer", Value: "standard_lowercase"},
							{Key: "searchAnalyzer", Value: "standard_lowercase"},
						}},
						{Key: "tags", Value: bson.D{
							{Key: "type", Value: "document"},
							{Key: "fields", Value: bson.D{
								{Key: "name", Value: bson.D{
									{Key: "type", Value: "string"},
									{Key: "analyzer", Value: "standard_lowercase"},
									{Key: "searchAnalyzer", Value: "standard_lowercase"},
								}},
							}},
						}},
						{Key: "isPrivate", Value: bson.D{
							{Key: "type", Value: "boolean"},
						}},
						{Key: "status", Value: bson.D{
							{Key: "type", Value: "string"},
						}},
						{Key: "settings", Value: bson.D{
							{Key: "type", Value: "document"},
							{Key: "fields", Value: bson.D{
								{Key: "showInDiscovery", Value: bson.D{
									{Key: "type", Value: "boolean"},
								}},
							}},
						}},
						{Key: "memberCount", Value: bson.D{
							{Key: "type", Value: "number"},
						}},
						{Key: "analytics", Value: bson.D{
							{Key: "type", Value: "document"},
							{Key: "fields", Value: bson.D{
								{Key: "trendingScore", Value: bson.D{
									{Key: "type", Value: "number"},
								}},
								{Key: "engagementScore", Value: bson.D{
									{Key: "type", Value: "number"},
								}},
								{Key: "qualityScore", Value: bson.D{
									{Key: "type", Value: "number"},
								}},
							}},
						}},
					}},
				}},
				{Key: "analyzers", Value: bson.A{
					bson.D{
						{Key: "name", Value: "standard_lowercase"},
						{Key: "charFilters", Value: bson.A{}},
						{Key: "tokenizer", Value: bson.D{
							{Key: "type", Value: "standard"},
						}},
						{Key: "tokenFilters", Value: bson.A{
							bson.D{
								{Key: "type", Value: "lowercase"},
							},
							bson.D{
								{Key: "type", Value: "edgeGram"},
								{Key: "minGram", Value: 2},
								{Key: "maxGram", Value: 20},
							},
						}},
					},
				}},
			},
			Options: options.SearchIndexes().SetName("community_search"),
		},
	}
	mongo.NewQueryBuilder[Community](db, CommunityCollectionName).Query(context.Background()).CheckSearchIndexes(searchIndexes)
}
