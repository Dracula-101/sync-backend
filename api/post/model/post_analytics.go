package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostAnalytics contains comprehensive analytics data for a post
type PostAnalytics struct {
	// Basic engagement metrics (auto-incremented)
	TotalViews     int64 `bson:"totalViews" json:"totalViews"`
	UniqueViews    int64 `bson:"uniqueViews" json:"uniqueViews"`
	TotalUpvotes   int64 `bson:"totalUpvotes" json:"totalUpvotes"`
	TotalDownvotes int64 `bson:"totalDownvotes" json:"totalDownvotes"`
	TotalComments  int64 `bson:"totalComments" json:"totalComments"`
	TotalShares    int64 `bson:"totalShares" json:"totalShares"`
	TotalSaves     int64 `bson:"totalSaves" json:"totalSaves"`
	TotalReports   int64 `bson:"totalReports" json:"totalReports"`
	TotalClicks    int64 `bson:"totalClicks" json:"totalClicks"`

	// Time-decay weighted metrics (auto-calculated for trending)
	WeightedViews1h  float64 `bson:"weightedViews1h" json:"weightedViews1h"`   // Views with 1-hour decay
	WeightedViews6h  float64 `bson:"weightedViews6h" json:"weightedViews6h"`   // Views with 6-hour decay
	WeightedViews24h float64 `bson:"weightedViews24h" json:"weightedViews24h"` // Views with 24-hour decay
	WeightedViews7d  float64 `bson:"weightedViews7d" json:"weightedViews7d"`   // Views with 7-day decay

	WeightedEngagement1h  float64 `bson:"weightedEngagement1h" json:"weightedEngagement1h"`   // Engagement with 1-hour decay
	WeightedEngagement6h  float64 `bson:"weightedEngagement6h" json:"weightedEngagement6h"`   // Engagement with 6-hour decay
	WeightedEngagement24h float64 `bson:"weightedEngagement24h" json:"weightedEngagement24h"` // Engagement with 24-hour decay
	WeightedEngagement7d  float64 `bson:"weightedEngagement7d" json:"weightedEngagement7d"`   // Engagement with 7-day decay

	// Velocity metrics (rate of change - auto-calculated)
	ViewVelocity1h       float64 `bson:"viewVelocity1h" json:"viewVelocity1h"`             // Views per hour
	ViewVelocity6h       float64 `bson:"viewVelocity6h" json:"viewVelocity6h"`             // Views per 6 hours
	EngagementVelocity1h float64 `bson:"engagementVelocity1h" json:"engagementVelocity1h"` // Engagement per hour
	EngagementVelocity6h float64 `bson:"engagementVelocity6h" json:"engagementVelocity6h"` // Engagement per 6 hours
	CommentVelocity1h    float64 `bson:"commentVelocity1h" json:"commentVelocity1h"`       // Comments per hour
	ShareVelocity1h      float64 `bson:"shareVelocity1h" json:"shareVelocity1h"`           // Shares per hour

	// Momentum indicators (acceleration - auto-calculated)
	ViewMomentum       float64 `bson:"viewMomentum" json:"viewMomentum"`             // View acceleration
	EngagementMomentum float64 `bson:"engagementMomentum" json:"engagementMomentum"` // Engagement acceleration
	TrendingMomentum   float64 `bson:"trendingMomentum" json:"trendingMomentum"`     // Overall trending momentum

	// Quality indicators (auto-calculated ratios)
	EngagementRate     float64 `bson:"engagementRate" json:"engagementRate"`         // Total engagement / views
	CommentToViewRatio float64 `bson:"commentToViewRatio" json:"commentToViewRatio"` // Comments / views
	ShareToViewRatio   float64 `bson:"shareToViewRatio" json:"shareToViewRatio"`     // Shares / views (virality)
	SaveToViewRatio    float64 `bson:"saveToViewRatio" json:"saveToViewRatio"`       // Saves / views
	UpvoteRatio        float64 `bson:"upvoteRatio" json:"upvoteRatio"`               // Upvotes / (upvotes + downvotes)
	ControversyScore   float64 `bson:"controversyScore" json:"controversyScore"`     // Measure of up/downvote balance

	// Composite scores (auto-calculated from multiple factors)
	HotScore        float64 `bson:"hotScore" json:"hotScore"`               // Reddit-style hot algorithm
	TrendingScore   float64 `bson:"trendingScore" json:"trendingScore"`     // Trending algorithm score
	QualityScore    float64 `bson:"qualityScore" json:"qualityScore"`       // Content quality score
	ViralityScore   float64 `bson:"viralityScore" json:"viralityScore"`     // Viral potential score
	PopularityScore float64 `bson:"popularityScore" json:"popularityScore"` // Long-term popularity
	RisingScore     float64 `bson:"risingScore" json:"risingScore"`         // Rising/emerging score

	// Time-bucketed activity for real-time updates
	ActivityBuckets PostActivityBuckets `bson:"activityBuckets" json:"activityBuckets"`

	// Age-weighted factors (auto-calculated based on post age)
	AgeInHours     float64 `bson:"ageInHours" json:"ageInHours"`         // Hours since creation
	AgePenalty     float64 `bson:"agePenalty" json:"agePenalty"`         // Age-based penalty factor
	FreshnessBoost float64 `bson:"freshnessBoost" json:"freshnessBoost"` // Freshness bonus

	// Performance tracking
	PeakViewsHour      *primitive.DateTime `bson:"peakViewsHour,omitempty" json:"peakViewsHour,omitempty"`
	PeakEngagementHour *primitive.DateTime `bson:"peakEngagementHour,omitempty" json:"peakEngagementHour,omitempty"`
	FirstViewAt        *primitive.DateTime `bson:"firstViewAt,omitempty" json:"firstViewAt,omitempty"`
	LastViewAt         *primitive.DateTime `bson:"lastViewAt,omitempty" json:"lastViewAt,omitempty"`
	LastEngagementAt   *primitive.DateTime `bson:"lastEngagementAt,omitempty" json:"lastEngagementAt,omitempty"`
	LastScoreUpdateAt  *primitive.DateTime `bson:"lastScoreUpdateAt,omitempty" json:"lastScoreUpdateAt,omitempty"`

	// Community contribution metrics
	CommunityEngagementContribution float64 `bson:"communityEngagementContribution" json:"communityEngagementContribution"` // How much this post contributes to community engagement
	CommunityActivityContribution   float64 `bson:"communityActivityContribution" json:"communityActivityContribution"`     // How much this post contributes to community activity
	CommunityGrowthContribution     float64 `bson:"communityGrowthContribution" json:"communityGrowthContribution"`         // How much this post contributes to community growth

	// User tracking for unique metrics
	UniqueViewers      []string         `bson:"uniqueViewers,omitempty" json:"-"`  // List of user IDs who viewed (for uniqueViews count)
	UniqueEngagers     []string         `bson:"uniqueEngagers,omitempty" json:"-"` // List of user IDs who engaged
	ViewerDemographics map[string]int64 `bson:"viewerDemographics,omitempty" json:"viewerDemographics,omitempty"`
}

// PostActivityBuckets tracks activity in different time periods for velocity calculations
type PostActivityBuckets struct {
	// Current time windows (sliding windows)
	CurrentHour    PostTimeWindowMetrics `bson:"currentHour" json:"currentHour"`
	Current6Hours  PostTimeWindowMetrics `bson:"current6Hours" json:"current6Hours"`
	Current24Hours PostTimeWindowMetrics `bson:"current24Hours" json:"current24Hours"`
	Current7Days   PostTimeWindowMetrics `bson:"current7Days" json:"current7Days"`

	// Previous time windows for momentum calculation
	PreviousHour    PostTimeWindowMetrics `bson:"previousHour" json:"previousHour"`
	Previous6Hours  PostTimeWindowMetrics `bson:"previous6Hours" json:"previous6Hours"`
	Previous24Hours PostTimeWindowMetrics `bson:"previous24Hours" json:"previous24Hours"`
	Previous7Days   PostTimeWindowMetrics `bson:"previous7Days" json:"previous7Days"`
}

// PostTimeWindowMetrics represents metrics for a specific time window
type PostTimeWindowMetrics struct {
	Views       int64              `bson:"views" json:"views"`
	UniqueViews int64              `bson:"uniqueViews" json:"uniqueViews"`
	Upvotes     int64              `bson:"upvotes" json:"upvotes"`
	Downvotes   int64              `bson:"downvotes" json:"downvotes"`
	Comments    int64              `bson:"comments" json:"comments"`
	Shares      int64              `bson:"shares" json:"shares"`
	Saves       int64              `bson:"saves" json:"saves"`
	Clicks      int64              `bson:"clicks" json:"clicks"`
	Reports     int64              `bson:"reports" json:"reports"`
	WindowStart primitive.DateTime `bson:"windowStart" json:"windowStart"`
	WindowEnd   primitive.DateTime `bson:"windowEnd" json:"windowEnd"`
	LastUpdated primitive.DateTime `bson:"lastUpdated" json:"lastUpdated"`
}

// PostInsights contains detailed analytics insights for a post
type PostInsights struct {
	PostId              string               `json:"postId"`
	Title               string               `json:"title"`
	TotalViews          int64                `json:"totalViews"`
	UniqueViews         int64                `json:"uniqueViews"`
	EngagementRate      float64              `json:"engagementRate"`
	ConversionRate      float64              `json:"conversionsRate"`
	ViralityScore       float64              `json:"viralityScore"`
	QualityScore        float64              `json:"qualityScore"`
	TrendingScore       float64              `json:"trendingScore"`
	HotScore            float64              `json:"hotScore"`
	RisingScore         float64              `json:"risingScore"`
	ViewVelocity        float64              `json:"viewVelocity"`
	EngagementVelocity  float64              `json:"engagementVelocity"`
	PeakActivity        *TimeSeriesData      `json:"peakActivity"`
	ViewSources         map[string]int64     `json:"viewSources"`
	Demographics        map[string]int64     `json:"demographics"`
	HourlyBreakdown     []*HourlyData        `json:"hourlyBreakdown"`
	DailyTrend          []*DailyData         `json:"dailyTrend"`
	EngagementBreakdown *EngagementBreakdown `json:"engagementBreakdown"`
	CompareToCommunity  *CommunityComparison `json:"compareToCommunity"`
	MomentumIndicators  *MomentumData        `json:"momentumIndicators"`
}

// PostComparisonData for comparing multiple posts
type PostComparisonData struct {
	PostId         string  `json:"postId"`
	Title          string  `json:"title"`
	Views          int64   `json:"views"`
	EngagementRate float64 `json:"engagementRate"`
	QualityScore   float64 `json:"qualityScore"`
	TrendingScore  float64 `json:"trendingScore"`
	HotScore       float64 `json:"hotScore"`
	ViewVelocity   float64 `json:"viewVelocity"`
	Age            int     `json:"ageInHours"`
	CommunityRank  int     `json:"communityRank"`
}

// AuthorPostStats contains analytics for all posts by an author
type AuthorPostStats struct {
	AuthorId         string   `json:"authorId"`
	TotalPosts       int64    `json:"totalPosts"`
	TotalViews       int64    `json:"totalViews"`
	TotalEngagement  int64    `json:"totalEngagement"`
	AvgQualityScore  float64  `json:"avgQualityScore"`
	AvgTrendingScore float64  `json:"avgTrendingScore"`
	BestPerforming   *Post    `json:"bestPerforming"`
	RecentTrend      string   `json:"recentTrend"`
	TopCommunities   []string `json:"topCommunities"`
}

// CommunityPostStats contains analytics for all posts in a community
type CommunityPostStats struct {
	CommunityId      string           `json:"communityId"`
	TotalPosts       int64            `json:"totalPosts"`
	TotalViews       int64            `json:"totalViews"`
	AvgEngagement    float64          `json:"avgEngagement"`
	AvgQuality       float64          `json:"avgQuality"`
	TopPerformers    []*Post          `json:"topPerformers"`
	TrendingPosts    []*Post          `json:"trendingPosts"`
	RisingPosts      []*Post          `json:"risingPosts"`
	PostsByType      map[string]int64 `json:"postsByType"`
	ActivityTrend    string           `json:"activityTrend"`
	GrowthIndicators *GrowthMetrics   `json:"growthIndicators"`
}

// PostSearchCriteria for advanced post searches
type PostSearchCriteria struct {
	MinViews         int64              `json:"minViews"`
	MaxViews         int64              `json:"maxViews"`
	MinEngagement    float64            `json:"minEngagement"`
	MaxEngagement    float64            `json:"maxEngagement"`
	MinQualityScore  float64            `json:"minQualityScore"`
	MinTrendingScore float64            `json:"minTrendingScore"`
	MinHotScore      float64            `json:"minHotScore"`
	MinViewVelocity  float64            `json:"minViewVelocity"`
	PostTypes        []string           `json:"postTypes"`
	Communities      []string           `json:"communities"`
	Authors          []string           `json:"authors"`
	Tags             []string           `json:"tags"`
	CreatedAfter     primitive.DateTime `json:"createdAfter"`
	CreatedBefore    primitive.DateTime `json:"createdBefore"`
	SortBy           string             `json:"sortBy"`
	SortOrder        string             `json:"sortOrder"`
	Limit            int                `json:"limit"`
}

// TimeSeriesData represents data points over time
type TimeSeriesData struct {
	Timestamp primitive.DateTime `json:"timestamp"`
	Value     int64              `json:"value"`
	Label     string             `json:"label,omitempty"`
}

// HourlyData represents hourly analytics data
type HourlyData struct {
	Hour     int   `json:"hour"`
	Views    int64 `json:"views"`
	Votes    int64 `json:"votes"`
	Comments int64 `json:"comments"`
	Shares   int64 `json:"shares"`
}

// DailyData represents daily analytics data
type DailyData struct {
	Date     primitive.DateTime `json:"date"`
	Views    int64              `json:"views"`
	Votes    int64              `json:"votes"`
	Comments int64              `json:"comments"`
	Shares   int64              `json:"shares"`
	Saves    int64              `json:"saves"`
}

// EngagementBreakdown shows different types of engagement
type EngagementBreakdown struct {
	Upvotes   int64 `json:"upvotes"`
	Downvotes int64 `json:"downvotes"`
	Comments  int64 `json:"comments"`
	Shares    int64 `json:"shares"`
	Saves     int64 `json:"saves"`
	Reports   int64 `json:"reports"`
}

// CommunityComparison compares post performance to community averages
type CommunityComparison struct {
	CommunityAvgViews      int64   `json:"communityAvgViews"`
	CommunityAvgEngagement float64 `json:"communityAvgEngagement"`
	CommunityAvgQuality    float64 `json:"communityAvgQuality"`
	PerformanceVsAvg       float64 `json:"performanceVsAvg"`
	Ranking                int     `json:"ranking"`
	TopPercentile          float64 `json:"topPercentile"`
}

// PostHeatmapData represents activity heatmap data
type PostHeatmapData struct {
	HourlyActivity map[int]int64    `json:"hourlyActivity"`
	DailyActivity  map[string]int64 `json:"dailyActivity"`
	WeeklyPattern  map[int]int64    `json:"weeklyPattern"`
}

// MomentumData represents momentum and velocity indicators
type MomentumData struct {
	ViewMomentum       float64 `json:"viewMomentum"`
	EngagementMomentum float64 `json:"engagementMomentum"`
	TrendingMomentum   float64 `json:"trendingMomentum"`
	Direction          string  `json:"direction"` // "up", "down", "stable"
	Acceleration       float64 `json:"acceleration"`
}

// GrowthMetrics represents growth indicators for communities
type GrowthMetrics struct {
	PostGrowthRate       float64 `json:"postGrowthRate"`
	EngagementGrowthRate float64 `json:"engagementGrowthRate"`
	QualityTrend         string  `json:"qualityTrend"`
	ActivityTrend        string  `json:"activityTrend"`
}

func NewPostTimeWindowMetrics() PostTimeWindowMetrics {
	return PostTimeWindowMetrics{
		Views:       0,
		UniqueViews: 0,
		Upvotes:     0,
		Downvotes:   0,
		Comments:    0,
		Shares:      0,
		Saves:       0,
		Clicks:      0,
		Reports:     0,
		WindowStart: primitive.NewDateTimeFromTime(time.Now()),
		WindowEnd:   primitive.NewDateTimeFromTime(time.Now().Add(1 * time.Hour)),
		LastUpdated: primitive.NewDateTimeFromTime(time.Now()),
	}
}

func NewPostActivityBuckets() PostActivityBuckets {
	return PostActivityBuckets{
		CurrentHour:    PostTimeWindowMetrics{},
		Current6Hours:  PostTimeWindowMetrics{},
		Current24Hours: PostTimeWindowMetrics{},
		Current7Days:   PostTimeWindowMetrics{},

		PreviousHour:    PostTimeWindowMetrics{},
		Previous6Hours:  PostTimeWindowMetrics{},
		Previous24Hours: PostTimeWindowMetrics{},
		Previous7Days:   PostTimeWindowMetrics{},
	}
}

func NewPostAnalytics() *PostAnalytics {
	return &PostAnalytics{
		TotalViews:     0,
		UniqueViews:    0,
		TotalUpvotes:   0,
		TotalDownvotes: 0,
		TotalComments:  0,
		TotalShares:    0,
		TotalSaves:     0,
		TotalReports:   0,
		TotalClicks:    0,

		WeightedViews1h:  0.0,
		WeightedViews6h:  0.0,
		WeightedViews24h: 0.0,
		WeightedViews7d:  0.0,

		WeightedEngagement1h:  0.0,
		WeightedEngagement6h:  0.0,
		WeightedEngagement24h: 0.0,
		WeightedEngagement7d:  0.0,

		ViewVelocity1h:       0.0,
		ViewVelocity6h:       0.0,
		EngagementVelocity1h: 0.0,
		EngagementVelocity6h: 0.0,
		CommentVelocity1h:    0.0,
		ShareVelocity1h:      0.0,

		ViewMomentum:       0.0,
		EngagementMomentum: 0.0,
		TrendingMomentum:   0.0,

		HotScore:      0.0,
		TrendingScore: 0.0,
		RisingScore:   0.0,
		ViralityScore: 0.0,
		QualityScore:  0.0,

		EngagementRate:     1.00, // Default to neutral engagement rate
		CommentToViewRatio: -1,   // Default to no comments
		ShareToViewRatio:   -1,   // Default to no shares
		SaveToViewRatio:    -1,   // Default to no saves
		UpvoteRatio:        -1,   // Default to no votes
		ControversyScore:   -1,   // Default to no votes

		ActivityBuckets: NewPostActivityBuckets(),
		AgeInHours:      0.0,
		AgePenalty:      0.0,
		PopularityScore: 0.0,
		FreshnessBoost:  0.0,

		CommunityEngagementContribution: 0.0,
		CommunityActivityContribution:   0.0,
		CommunityGrowthContribution:     0.0,
		UniqueViewers:                   []string{},

		UniqueEngagers:     []string{},
		ViewerDemographics: make(map[string]int64),
	}
}
