package analytics

import (
	"context"
	"errors"
	"math"
	"sync-backend/api/community/model"
	"sync-backend/arch/mongo"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CommunityAnalytics defines the contract for community analytics operations
type CommunityAnalytics interface {
	// Activity tracking methods
	RecordCommunityView(communityId, userId string) error
	RecordPostCreated(communityId, userId string) error
	RecordCommentCreated(communityId, userId string) error
	RecordReaction(communityId, userId string, reactionType string) error
	RecordShare(communityId, userId string) error
	RecordMemberJoin(communityId, userId string) error
	RecordMemberLeave(communityId, userId string) error
	RecordReport(communityId, userId string) error

	// Analytics retrieval methods
	GetCommunityAnalytics(communityId string) (*model.Community, error)
	GetTrendingCommunities(limit int, timeframe string) ([]*model.Community, error)
	GetTopCommunitiesByEngagement(limit int) ([]*model.Community, error)
	GetTopCommunitiesByGrowth(limit int, period string) ([]*model.Community, error)
	GetTopCommunitiesByActivity(limit int, hours int) ([]*model.Community, error)
	GetCommunitiesByQualityScore(limit int) ([]*model.Community, error)

	// Batch operations
	UpdateActiveMembers(communityId string, activeUserIds []string) error
	BulkUpdateScores(communityIds []string) error

	// Analytics queries
	GetCommunityInsights(communityId string, days int) (*CommunityInsights, error)
	GetCommunityComparison(communityIds []string) ([]*CommunityComparisonData, error)
	SearchCommunitiesByScore(query SearchCriteria) ([]*model.Community, error)

	// Maintenance operations
	CleanupOldAnalytics(olderThanDays int) error
	RecalculateAllScores(ctx context.Context) error
}

// CommunityInsights represents detailed analytics for a community
type CommunityInsights struct {
	CommunityId         string              `json:"communityId"`
	Period              int                 `json:"period"` // days
	MemberGrowth        []DailyMetric       `json:"memberGrowth"`
	ActivityTrends      []DailyMetric       `json:"activityTrends"`
	EngagementBreakdown EngagementBreakdown `json:"engagementBreakdown"`
	PeakActivityHours   []HourlyActivity    `json:"peakActivityHours"`
	TopContributors     []ContributorStats  `json:"topContributors"`
	ContentStats        ContentStatistics   `json:"contentStats"`
	ComparisonToAverage ComparisonMetrics   `json:"comparisonToAverage"`
}

type DailyMetric struct {
	Date  time.Time `json:"date"`
	Value int64     `json:"value"`
}

type EngagementBreakdown struct {
	Posts             int64   `json:"posts"`
	Comments          int64   `json:"comments"`
	Likes             int64   `json:"likes"`
	Shares            int64   `json:"shares"`
	Views             int64   `json:"views"`
	AvgPostEngagement float64 `json:"avgPostEngagement"`
}

type HourlyActivity struct {
	Hour     int     `json:"hour"`
	Activity int64   `json:"activity"`
	Score    float64 `json:"score"`
}

type ContributorStats struct {
	UserId        string  `json:"userId"`
	Username      string  `json:"username"`
	PostCount     int64   `json:"postCount"`
	CommentCount  int64   `json:"commentCount"`
	LikesGiven    int64   `json:"likesGiven"`
	LikesReceived int64   `json:"likesReceived"`
	Score         float64 `json:"score"`
}

type ContentStatistics struct {
	TotalPosts         int64      `json:"totalPosts"`
	TotalComments      int64      `json:"totalComments"`
	AvgPostLength      float64    `json:"avgPostLength"`
	AvgCommentsPerPost float64    `json:"avgCommentsPerPost"`
	PopularTags        []TagStats `json:"popularTags"`
}

type TagStats struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

type ComparisonMetrics struct {
	EngagementVsAverage float64 `json:"engagementVsAverage"` // percentage
	GrowthVsAverage     float64 `json:"growthVsAverage"`
	ActivityVsAverage   float64 `json:"activityVsAverage"`
}

type CommunityComparisonData struct {
	CommunityId     string  `json:"communityId"`
	Name            string  `json:"name"`
	MemberCount     int64   `json:"memberCount"`
	EngagementScore float64 `json:"engagementScore"`
	TrendingScore   float64 `json:"trendingScore"`
	QualityScore    float64 `json:"qualityScore"`
	GrowthRate      float64 `json:"growthRate"`
}

type SearchCriteria struct {
	MinMembers     int64      `json:"minMembers"`
	MaxMembers     int64      `json:"maxMembers"`
	MinEngagement  float64    `json:"minEngagement"`
	MinTrending    float64    `json:"minTrending"`
	MinQuality     float64    `json:"minQuality"`
	CreatedAfter   *time.Time `json:"createdAfter"`
	Tags           []string   `json:"tags"`
	ExcludePrivate bool       `json:"excludePrivate"`
	IsActive       bool       `json:"isActive"`
	SortBy         string     `json:"sortBy"`    // "trending", "engagement", "quality", "members", "growth"
	SortOrder      string     `json:"sortOrder"` // "asc", "desc"
}

// communityAnalyticsService implements CommunityAnalyticsInterface
type communityAnalyticsService struct {
	communityQB mongo.QueryBuilder[model.Community]
	ctx         context.Context
}

func NewCommunityAnalyticsService(db mongo.Database) CommunityAnalytics {
	return &communityAnalyticsService{
		communityQB: mongo.NewQueryBuilder[model.Community](db, model.CommunityCollectionName),
		ctx:         context.Background(),
	}
}

// Activity tracking methods
func (s *communityAnalyticsService) RecordCommunityView(communityId, userId string) error {
	update := bson.M{
		"$inc": bson.M{
			"analytics.totalViews":                        1,
			"analytics.activityBuckets.currentHour.views": 1,
		},
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
			"updatedAt":                primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	return s.updateCommunityAnalytics(communityId, update)
}

func (s *communityAnalyticsService) RecordPostCreated(communityId string, userId string) error {
	now := time.Now()
	update := bson.M{
		"$inc": bson.M{
			"postCount": 1,
			"analytics.activityBuckets.currentHour.posts": 1,
		},
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(now),
			"analytics.lastPostAt":     primitive.NewDateTimeFromTime(now),
			"updatedAt":                primitive.NewDateTimeFromTime(now),
		},
	}

	if err := s.updateCommunityAnalytics(communityId, update); err != nil {
		return err
	}

	// Trigger score recalculation for significant events
	return s.maybeUpdateScores(communityId, "post")
}

func (s *communityAnalyticsService) RecordCommentCreated(communityId, userId string) error {
	now := time.Now()
	update := bson.M{
		"$inc": bson.M{
			"analytics.totalComments":                        1,
			"analytics.activityBuckets.currentHour.comments": 1,
		},
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(now),
			"analytics.lastCommentAt":  primitive.NewDateTimeFromTime(now),
			"updatedAt":                primitive.NewDateTimeFromTime(now),
		},
	}

	if err := s.updateCommunityAnalytics(communityId, update); err != nil {
		return err
	}

	return s.maybeUpdateScores(communityId, "comment")
}

func (s *communityAnalyticsService) RecordReaction(communityId, userId string, reactionType string) error {
	update := bson.M{
		"$inc": bson.M{
			"analytics.totalLikes":                        1,
			"analytics.activityBuckets.currentHour.likes": 1,
		},
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
			"updatedAt":                primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	return s.updateCommunityAnalytics(communityId, update)
}

func (s *communityAnalyticsService) RecordShare(communityId, userId string) error {
	update := bson.M{
		"$inc": bson.M{
			"analytics.totalShares":                        1,
			"analytics.activityBuckets.currentHour.shares": 1,
		},
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
			"updatedAt":                primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	return s.updateCommunityAnalytics(communityId, update)
}

func (s *communityAnalyticsService) RecordMemberJoin(communityId, userId string) error {
	now := time.Now()
	update := bson.M{
		"$inc": bson.M{
			"memberCount":                                      1,
			"analytics.memberJoinsToday":                       1,
			"analytics.memberJoinsWeek":                        1,
			"analytics.memberJoinsMonth":                       1,
			"analytics.activeMembersToday":                     1,
			"analytics.activeMembersWeek":                      1,
			"analytics.activeMembersMonth":                     1,
			"analytics.activityBuckets.currentHour.newMembers": 1,
		},
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(now),
			"updatedAt":                primitive.NewDateTimeFromTime(now),
		},
	}

	if err := s.updateCommunityAnalytics(communityId, update); err != nil {
		return err
	}

	return s.maybeUpdateScores(communityId, "join")
}

func (s *communityAnalyticsService) RecordMemberLeave(communityId, userId string) error {
	update := bson.M{
		"$inc": bson.M{
			"memberCount": -1,
		},
		"$set": bson.M{
			"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if err := s.updateCommunityAnalytics(communityId, update); err != nil {
		return err
	}

	return s.maybeUpdateScores(communityId, "leave")
}

func (s *communityAnalyticsService) RecordReport(communityId, userId string) error {
	// Reports might negatively impact quality score
	update := bson.M{
		"$set": bson.M{
			"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if err := s.updateCommunityAnalytics(communityId, update); err != nil {
		return err
	}

	return s.maybeUpdateScores(communityId, "report")
}

// Analytics retrieval methods
func (s *communityAnalyticsService) GetCommunityAnalytics(communityId string) (*model.Community, error) {
	filter := bson.M{"communityId": communityId}
	result, err := s.communityQB.Query(s.ctx).FindOne(filter, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *communityAnalyticsService) GetTrendingCommunities(limit int, timeframe string) ([]*model.Community, error) {
	// Adjust trending score calculation based on timeframe
	var filter bson.M
	switch timeframe {
	case "hour":
		filter = bson.M{
			"status":                   model.CommunityStatusActive,
			"settings.showInDiscovery": true,
			"isPrivate":                false,
		}
	case "day", "week", "month":
		filter = bson.M{
			"status":                   model.CommunityStatusActive,
			"settings.showInDiscovery": true,
			"isPrivate":                false,
		}
	default:
		return nil, errors.New("invalid timeframe: use 'hour', 'day', 'week', or 'month'")
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "analytics.trendingScore", Value: -1}}).
		SetLimit(int64(limit))

	return s.communityQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *communityAnalyticsService) GetTopCommunitiesByEngagement(limit int) ([]*model.Community, error) {
	filter := bson.M{
		"status":                   model.CommunityStatusActive,
		"settings.showInDiscovery": true,
		"isPrivate":                false,
		"memberCount":              bson.M{"$gte": 5}, // Minimum members for meaningful engagement
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "analytics.engagementScore", Value: -1}}).
		SetLimit(int64(limit))

	return s.communityQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *communityAnalyticsService) GetTopCommunitiesByGrowth(limit int, period string) ([]*model.Community, error) {
	var growthField string
	switch period {
	case "day":
		growthField = "analytics.memberJoinsToday"
	case "week":
		growthField = "analytics.memberJoinsWeek"
	case "month":
		growthField = "analytics.memberJoinsMonth"
	default:
		return nil, errors.New("invalid period: use 'day', 'week', or 'month'")
	}

	filter := bson.M{
		"status":                   model.CommunityStatusActive,
		"settings.showInDiscovery": true,
		"isPrivate":                false,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: growthField, Value: -1}}).
		SetLimit(int64(limit))

	return s.communityQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *communityAnalyticsService) GetTopCommunitiesByActivity(limit int, hours int) ([]*model.Community, error) {
	// Calculate activity score based on recent hours
	hoursAgo := time.Now().Add(-time.Duration(hours) * time.Hour)

	filter := bson.M{
		"status":                   model.CommunityStatusActive,
		"settings.showInDiscovery": true,
		"isPrivate":                false,
		"analytics.lastActivityAt": bson.M{"$gte": primitive.NewDateTimeFromTime(hoursAgo)},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "analytics.lastActivityAt", Value: -1}}).
		SetLimit(int64(limit))

	return s.communityQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *communityAnalyticsService) GetCommunitiesByQualityScore(limit int) ([]*model.Community, error) {
	filter := bson.M{
		"status":                   model.CommunityStatusActive,
		"settings.showInDiscovery": true,
		"isPrivate":                false,
		"memberCount":              bson.M{"$gte": 10}, // Minimum members for quality assessment
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "analytics.qualityScore", Value: -1}}).
		SetLimit(int64(limit))

	return s.communityQB.Query(s.ctx).FindAll(filter, opts)
}

// Batch operations
func (s *communityAnalyticsService) UpdateActiveMembers(communityId string, activeUserIds []string) error {
	activeCount := int64(len(activeUserIds))

	update := bson.M{
		"$set": bson.M{
			"analytics.activeMembersToday": activeCount,
			"updatedAt":                    primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	return s.updateCommunityAnalytics(communityId, update)
}

func (s *communityAnalyticsService) BulkUpdateScores(communityIds []string) error {
	for _, communityId := range communityIds {
		if err := s.recalculateScores(communityId); err != nil {
			// Log error but continue with other communities
			continue
		}
	}
	return nil
}

// Analytics queries
func (s *communityAnalyticsService) GetCommunityInsights(communityId string, days int) (*CommunityInsights, error) {
	community, err := s.GetCommunityAnalytics(communityId)
	if err != nil {
		return nil, err
	}

	// Build insights from community data and activity buckets
	insights := &CommunityInsights{
		CommunityId: communityId,
		Period:      days,
		EngagementBreakdown: EngagementBreakdown{
			Posts:    community.PostCount,
			Comments: community.Analytics.TotalComments,
			Likes:    community.Analytics.TotalLikes,
			Shares:   community.Analytics.TotalShares,
			Views:    community.Analytics.TotalViews,
		},
		ContentStats: ContentStatistics{
			TotalPosts:    community.PostCount,
			TotalComments: community.Analytics.TotalComments,
		},
	}

	// Calculate averages
	if community.PostCount > 0 {
		insights.EngagementBreakdown.AvgPostEngagement = float64(community.Analytics.TotalLikes+community.Analytics.TotalComments) / float64(community.PostCount)
		insights.ContentStats.AvgCommentsPerPost = float64(community.Analytics.TotalComments) / float64(community.PostCount)
	}

	return insights, nil
}

func (s *communityAnalyticsService) GetCommunityComparison(communityIds []string) ([]*CommunityComparisonData, error) {
	filter := bson.M{"communityId": bson.M{"$in": communityIds}}
	communities, err := s.communityQB.Query(s.ctx).FindAll(filter, nil)
	if err != nil {
		return nil, err
	}

	var comparisons []*CommunityComparisonData
	for _, community := range communities {
		comparison := &CommunityComparisonData{
			CommunityId:     community.CommunityId,
			Name:            community.Name,
			MemberCount:     community.MemberCount,
			EngagementScore: community.Analytics.EngagementScore,
			TrendingScore:   community.Analytics.TrendingScore,
			QualityScore:    community.Analytics.QualityScore,
		}

		// Calculate growth rate (joins this week / total members)
		if community.MemberCount > 0 {
			comparison.GrowthRate = (float64(community.Analytics.MemberJoinsWeek) / float64(community.MemberCount)) * 100
		}

		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

func (s *communityAnalyticsService) SearchCommunitiesByScore(criteria SearchCriteria) ([]*model.Community, error) {
	filter := s.buildSearchFilter(criteria)
	opts := s.buildSearchOptions(criteria)

	return s.communityQB.Query(s.ctx).FindAll(filter, opts)
}

// Maintenance operations
func (s *communityAnalyticsService) CleanupOldAnalytics(olderThanDays int) error {
	// This would typically involve cleaning up detailed activity buckets
	// For now, we'll just ensure communities older than specified days have their buckets reset
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	filter := bson.M{
		"createdAt": bson.M{"$lt": primitive.NewDateTimeFromTime(cutoffDate)},
	}

	update := bson.M{
		"$set": bson.M{
			"analytics.activityBuckets.last30Days": make([]model.ActivityBucket, 30),
		},
	}

	_, err := s.communityQB.Query(s.ctx).UpdateMany(filter, update, nil)
	return err
}

func (s *communityAnalyticsService) RecalculateAllScores(ctx context.Context) error {
	// Get all active communities
	filter := bson.M{"status": model.CommunityStatusActive}
	communities, err := s.communityQB.Query(s.ctx).FindAll(filter, nil)
	if err != nil {
		return err
	}

	for _, community := range communities {
		if err := s.recalculateScores(community.CommunityId); err != nil {
			// Log error but continue
			continue
		}
	}

	return nil
}

// Helper methods
func (s *communityAnalyticsService) updateCommunityAnalytics(communityId string, update bson.M) error {
	filter := bson.M{"communityId": communityId}
	_, err := s.communityQB.Query(s.ctx).UpdateOne(filter, update, nil)
	return err
}

func (s *communityAnalyticsService) maybeUpdateScores(communityId string, eventType string) error {
	// Update scores for significant events or periodically
	significantEvents := map[string]bool{
		"post":  true,
		"join":  true,
		"leave": true,
	}

	if significantEvents[eventType] {
		return s.recalculateScores(communityId)
	}

	return nil
}

func (s *communityAnalyticsService) recalculateScores(communityId string) error {
	community, err := s.GetCommunityAnalytics(communityId)
	if err != nil {
		return err
	}

	now := time.Now()

	// Calculate engagement score
	var engagementScore float64
	if community.MemberCount > 0 {
		totalEngagement := community.Analytics.TotalLikes + community.Analytics.TotalComments + community.Analytics.TotalShares
		engagementRate := float64(totalEngagement) / float64(community.MemberCount)
		engagementScore = math.Min(engagementRate*10, 100)
	}

	// Calculate trending score (based on recent activity)
	var trendingScore float64
	recentActivity := s.calculateRecentActivityScore(community)
	communityAge := now.Sub(community.CreatedAt.Time()).Hours() / 24
	ageBonus := math.Max(1.0, 30.0/math.Max(communityAge, 1.0))
	trendingScore = recentActivity * ageBonus

	// Calculate quality score
	var qualityScore float64
	if community.MemberCount > 0 {
		memberRetention := float64(community.Analytics.ActiveMembersWeek) / float64(community.MemberCount)
		var postQuality float64
		if community.PostCount > 0 {
			postQuality = float64(community.Analytics.TotalLikes) / float64(community.PostCount)
		}
		qualityScore = math.Min((memberRetention*50)+(postQuality*5), 100)
	}

	// Update scores
	update := bson.M{
		"$set": bson.M{
			"analytics.engagementScore": engagementScore,
			"analytics.trendingScore":   trendingScore,
			"analytics.qualityScore":    qualityScore,
			"analytics.scoresUpdatedAt": primitive.NewDateTimeFromTime(now),
			"updatedAt":                 primitive.NewDateTimeFromTime(now),
		},
	}

	return s.updateCommunityAnalytics(communityId, update)
}

func (s *communityAnalyticsService) calculateRecentActivityScore(community *model.Community) float64 {
	var totalActivity int64

	// Calculate from activity buckets
	for _, bucket := range community.Analytics.ActivityBuckets.Last24Hours {
		totalActivity += bucket.Posts*5 + bucket.Comments*2 + bucket.Likes + bucket.Views/10 + bucket.Shares*3
	}

	// Add current hour
	currentHour := community.Analytics.ActivityBuckets.CurrentHour
	totalActivity += currentHour.Posts*5 + currentHour.Comments*2 + currentHour.Likes + currentHour.Views/10 + currentHour.Shares*3

	return math.Min(float64(totalActivity)/10, 100)
}

func (s *communityAnalyticsService) buildSearchFilter(criteria SearchCriteria) bson.M {
	filter := bson.M{}

	if criteria.IsActive {
		filter["status"] = model.CommunityStatusActive
	}

	if criteria.ExcludePrivate {
		filter["isPrivate"] = false
		filter["settings.showInDiscovery"] = true
	}

	if criteria.MinMembers > 0 || criteria.MaxMembers > 0 {
		memberFilter := bson.M{}
		if criteria.MinMembers > 0 {
			memberFilter["$gte"] = criteria.MinMembers
		}
		if criteria.MaxMembers > 0 {
			memberFilter["$lte"] = criteria.MaxMembers
		}
		filter["memberCount"] = memberFilter
	}

	if criteria.MinEngagement > 0 {
		filter["analytics.engagementScore"] = bson.M{"$gte": criteria.MinEngagement}
	}

	if criteria.MinTrending > 0 {
		filter["analytics.trendingScore"] = bson.M{"$gte": criteria.MinTrending}
	}

	if criteria.MinQuality > 0 {
		filter["analytics.qualityScore"] = bson.M{"$gte": criteria.MinQuality}
	}

	if criteria.CreatedAfter != nil {
		filter["createdAt"] = bson.M{"$gte": primitive.NewDateTimeFromTime(*criteria.CreatedAfter)}
	}

	if len(criteria.Tags) > 0 {
		filter["tags.name"] = bson.M{"$in": criteria.Tags}
	}

	return filter
}

func (s *communityAnalyticsService) buildSearchOptions(criteria SearchCriteria) *options.FindOptions {
	opts := options.Find()

	// Build sort
	var sortField string
	var sortValue int = -1 // Default descending

	switch criteria.SortBy {
	case "trending":
		sortField = "analytics.trendingScore"
	case "engagement":
		sortField = "analytics.engagementScore"
	case "quality":
		sortField = "analytics.qualityScore"
	case "members":
		sortField = "memberCount"
	case "growth":
		sortField = "analytics.memberJoinsWeek"
	default:
		sortField = "analytics.trendingScore"
	}

	if criteria.SortOrder == "asc" {
		sortValue = 1
	}

	opts.SetSort(bson.D{{Key: sortField, Value: sortValue}})

	return opts
}
