package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"sync-backend/api/post/model"
	"sync-backend/arch/mongo"
)

type PostAnalytics interface {
	RecordPostView(postId, userId string) error
	RecordPostClick(postId, userId string) error
	RecordPostVote(postId, userId string, voteType model.VoteType) error
	RecordPostComment(postId, userId string) error
	RecordPostShare(postId, userId string) error
	RecordPostSave(postId, userId string) error
	RecordPostReport(postId, userId string) error

	GetPostAnalytics(postId string) (*model.Post, error)
	GetTrendingPosts(limit int, timeframe string, communityId string) ([]*model.Post, error)
	GetHotPosts(limit int, communityId string) ([]*model.Post, error)
	GetRisingPosts(limit int, communityId string) ([]*model.Post, error)
	GetTopPostsByEngagement(limit int, communityId string) ([]*model.Post, error)
	GetTopPostsByViews(limit int, timeframe string, communityId string) ([]*model.Post, error)
	GetTopPostsByQuality(limit int, communityId string) ([]*model.Post, error)
	GetViralPosts(limit int, minViralityScore float64, communityId string) ([]*model.Post, error)
	GetControversialPosts(limit int, communityId string) ([]*model.Post, error)

	GetCommunityTopPosts(communityId string, sortBy string, limit int) ([]*model.Post, error)
	GetCommunityTrendingAnalytics(communityId string) (*model.CommunityPostStats, error)
	GetCommunityEngagementLeaders(communityId string, limit int) ([]*model.Post, error)
	GetCommunityRisingStars(communityId string, limit int) ([]*model.Post, error)

	GetPostInsights(postId string, days int) (*model.PostInsights, error)
	GetPostPerformanceComparison(postIds []string) ([]*model.PostComparisonData, error)
	GetAuthorPostAnalytics(authorId string, days int) (*model.AuthorPostStats, error)
	SearchPostsByMetrics(criteria model.PostSearchCriteria) ([]*model.Post, error)

	CalculateAndUpdatePostScores(postId string) error
	GetPostsRequiringScoreUpdate() ([]*model.Post, error)

	CleanupOldPostAnalytics(olderThanDays int) error
	ArchivePostAnalytics(postIds []string) error
}

type postAnalyticsService struct {
	postQB mongo.QueryBuilder[model.Post]
	ctx    context.Context
}

func NewPostAnalyticsService(db mongo.Database) PostAnalytics {
	return &postAnalyticsService{
		postQB: mongo.NewQueryBuilder[model.Post](db, model.PostCollectionName),
		ctx:    context.Background(),
	}
}

func (s *postAnalyticsService) RecordPostView(postId, userId string) error {
	filter := bson.M{"postId": postId}

	update := bson.M{
		"$inc": bson.M{
			"viewCount":            1,
			"analytics.totalViews": 1,
			"analytics.activityBuckets.currentHour.views":    1,
			"analytics.activityBuckets.current6Hours.views":  1,
			"analytics.activityBuckets.current24Hours.views": 1,
			"analytics.activityBuckets.current7Days.views":   1,
		},
		"$set": bson.M{
			"updatedAt":            primitive.NewDateTimeFromTime(time.Now()),
			"lastActivityAt":       primitive.NewDateTimeFromTime(time.Now()),
			"analytics.lastViewAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if userId != "" {
		update["$addToSet"] = bson.M{
			"analytics.uniqueViewers": userId,
		}
		update["$inc"].(bson.M)["analytics.uniqueViews"] = 1
		update["$inc"].(bson.M)["analytics.activityBuckets.currentHour.uniqueViews"] = 1
		update["$inc"].(bson.M)["analytics.activityBuckets.current6Hours.uniqueViews"] = 1
		update["$inc"].(bson.M)["analytics.activityBuckets.current24Hours.uniqueViews"] = 1
		update["$inc"].(bson.M)["analytics.activityBuckets.current7Days.uniqueViews"] = 1
	}

	_, err := s.postQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdatePostScores(postId)
}

func (s *postAnalyticsService) RecordPostVote(postId, userId string, voteType model.VoteType) error {
	filter := bson.M{"postId": postId}

	update := bson.M{
		"$set": bson.M{
			fmt.Sprintf("voters.%s", userId): voteType,
			"updatedAt":                      primitive.NewDateTimeFromTime(time.Now()),
			"lastActivityAt":                 primitive.NewDateTimeFromTime(time.Now()),
			"analytics.lastEngagementAt":     primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if voteType == model.Upvote {
		update["$inc"] = bson.M{
			"synergy":                1,
			"analytics.totalUpvotes": 1,
			"analytics.activityBuckets.currentHour.upvotes":    1,
			"analytics.activityBuckets.current6Hours.upvotes":  1,
			"analytics.activityBuckets.current24Hours.upvotes": 1,
			"analytics.activityBuckets.current7Days.upvotes":   1,
		}
	} else if voteType == model.Downvote {
		update["$inc"] = bson.M{
			"synergy":                  -1,
			"analytics.totalDownvotes": 1,
			"analytics.activityBuckets.currentHour.downvotes":    1,
			"analytics.activityBuckets.current6Hours.downvotes":  1,
			"analytics.activityBuckets.current24Hours.downvotes": 1,
			"analytics.activityBuckets.current7Days.downvotes":   1,
		}
	}

	if userId != "" {
		update["$addToSet"] = bson.M{
			"analytics.uniqueEngagers": userId,
		}
	}

	_, err := s.postQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdatePostScores(postId)
}

func (s *postAnalyticsService) RecordPostComment(postId, userId string) error {
	filter := bson.M{"postId": postId}

	update := bson.M{
		"$inc": bson.M{
			"commentCount":            1,
			"analytics.totalComments": 1,
			"analytics.activityBuckets.currentHour.comments":    1,
			"analytics.activityBuckets.current6Hours.comments":  1,
			"analytics.activityBuckets.current24Hours.comments": 1,
			"analytics.activityBuckets.current7Days.comments":   1,
		},
		"$set": bson.M{
			"updatedAt":                  primitive.NewDateTimeFromTime(time.Now()),
			"lastActivityAt":             primitive.NewDateTimeFromTime(time.Now()),
			"analytics.lastEngagementAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if userId != "" {
		update["$addToSet"] = bson.M{
			"analytics.uniqueEngagers": userId,
		}
	}

	_, err := s.postQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdatePostScores(postId)
}

func (s *postAnalyticsService) RecordPostShare(postId, userId string) error {
	filter := bson.M{"postId": postId}

	update := bson.M{
		"$inc": bson.M{
			"shareCount":            1,
			"analytics.totalShares": 1,
			"analytics.activityBuckets.currentHour.shares":    1,
			"analytics.activityBuckets.current6Hours.shares":  1,
			"analytics.activityBuckets.current24Hours.shares": 1,
			"analytics.activityBuckets.current7Days.shares":   1,
		},
		"$set": bson.M{
			"updatedAt":                  primitive.NewDateTimeFromTime(time.Now()),
			"lastActivityAt":             primitive.NewDateTimeFromTime(time.Now()),
			"analytics.lastEngagementAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if userId != "" {
		update["$addToSet"] = bson.M{
			"analytics.uniqueEngagers": userId,
		}
	}

	_, err := s.postQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdatePostScores(postId)
}

func (s *postAnalyticsService) RecordPostSave(postId, userId string) error {
	filter := bson.M{"postId": postId}

	update := bson.M{
		"$inc": bson.M{
			"saveCount":            1,
			"analytics.totalSaves": 1,
			"analytics.activityBuckets.currentHour.saves":    1,
			"analytics.activityBuckets.current6Hours.saves":  1,
			"analytics.activityBuckets.current24Hours.saves": 1,
			"analytics.activityBuckets.current7Days.saves":   1,
		},
		"$set": bson.M{
			"updatedAt":                  primitive.NewDateTimeFromTime(time.Now()),
			"lastActivityAt":             primitive.NewDateTimeFromTime(time.Now()),
			"analytics.lastEngagementAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	if userId != "" {
		update["$addToSet"] = bson.M{
			"analytics.uniqueEngagers": userId,
		}
	}

	_, err := s.postQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdatePostScores(postId)
}

func (s *postAnalyticsService) RecordPostReport(postId, userId string) error {
	filter := bson.M{"postId": postId}

	update := bson.M{
		"$inc": bson.M{
			"analytics.totalReports":                           1,
			"analytics.activityBuckets.currentHour.reports":    1,
			"analytics.activityBuckets.current6Hours.reports":  1,
			"analytics.activityBuckets.current24Hours.reports": 1,
			"analytics.activityBuckets.current7Days.reports":   1,
		},
		"$set": bson.M{
			"updatedAt":      primitive.NewDateTimeFromTime(time.Now()),
			"lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := s.postQB.Query(s.ctx).UpdateOne(filter, update, nil)
	return err
}

func (s *postAnalyticsService) RecordPostClick(postId, userId string) error {
	filter := bson.M{"postId": postId}

	update := bson.M{
		"$inc": bson.M{
			"analytics.totalClicks":                           1,
			"analytics.activityBuckets.currentHour.clicks":    1,
			"analytics.activityBuckets.current6Hours.clicks":  1,
			"analytics.activityBuckets.current24Hours.clicks": 1,
			"analytics.activityBuckets.current7Days.clicks":   1,
		},
		"$set": bson.M{
			"updatedAt":      primitive.NewDateTimeFromTime(time.Now()),
			"lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := s.postQB.Query(s.ctx).UpdateOne(filter, update, nil)
	return err
}

func (s *postAnalyticsService) GetPostAnalytics(postId string) (*model.Post, error) {
	filter := bson.M{"postId": postId}
	return s.postQB.Query(s.ctx).FindOne(filter, nil)
}

func (s *postAnalyticsService) GetTrendingPosts(limit int, timeframe string, communityId string) ([]*model.Post, error) {
	filter := bson.M{
		"status": model.PostStatusActive,
	}

	if communityId != "" {
		filter["communityId"] = communityId
	}

	if timeframe != "" {
		var timeThreshold time.Time
		switch timeframe {
		case "1h":
			timeThreshold = time.Now().Add(-1 * time.Hour)
		case "6h":
			timeThreshold = time.Now().Add(-6 * time.Hour)
		case "24h":
			timeThreshold = time.Now().Add(-24 * time.Hour)
		case "7d":
			timeThreshold = time.Now().Add(-7 * 24 * time.Hour)
		}

		if !timeThreshold.IsZero() {
			filter["createdAt"] = bson.M{"$gte": primitive.NewDateTimeFromTime(timeThreshold)}
		}
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.trendingScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetHotPosts(limit int, communityId string) ([]*model.Post, error) {
	filter := bson.M{
		"status": model.PostStatusActive,
	}

	if communityId != "" {
		filter["communityId"] = communityId
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.hotScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetRisingPosts(limit int, communityId string) ([]*model.Post, error) {
	filter := bson.M{
		"status":    model.PostStatusActive,
		"createdAt": bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().Add(-48 * time.Hour))},
	}

	if communityId != "" {
		filter["communityId"] = communityId
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.risingScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetTopPostsByEngagement(limit int, communityId string) ([]*model.Post, error) {
	filter := bson.M{
		"status": model.PostStatusActive,
	}

	if communityId != "" {
		filter["communityId"] = communityId
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.engagementRate", Value: -1}},
		Limit: &limitVal,
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetTopPostsByViews(limit int, timeframe string, communityId string) ([]*model.Post, error) {
	filter := bson.M{
		"status": model.PostStatusActive,
	}

	if communityId != "" {
		filter["communityId"] = communityId
	}

	if timeframe != "" {
		var timeThreshold time.Time
		switch timeframe {
		case "1h":
			timeThreshold = time.Now().Add(-1 * time.Hour)
		case "6h":
			timeThreshold = time.Now().Add(-6 * time.Hour)
		case "24h":
			timeThreshold = time.Now().Add(-24 * time.Hour)
		case "7d":
			timeThreshold = time.Now().Add(-7 * 24 * time.Hour)
		}

		if !timeThreshold.IsZero() {
			filter["createdAt"] = bson.M{"$gte": primitive.NewDateTimeFromTime(timeThreshold)}
		}
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.totalViews", Value: -1}},
		Limit: &limitVal,
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetTopPostsByQuality(limit int, communityId string) ([]*model.Post, error) {
	filter := bson.M{
		"status":                 model.PostStatusActive,
		"analytics.qualityScore": bson.M{"$gt": 0},
	}

	if communityId != "" {
		filter["communityId"] = communityId
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.qualityScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetViralPosts(limit int, minViralityScore float64, communityId string) ([]*model.Post, error) {
	filter := bson.M{
		"status":                  model.PostStatusActive,
		"analytics.viralityScore": bson.M{"$gte": minViralityScore},
	}

	if communityId != "" {
		filter["communityId"] = communityId
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.viralityScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetControversialPosts(limit int, communityId string) ([]*model.Post, error) {
	filter := bson.M{
		"status":                     model.PostStatusActive,
		"analytics.controversyScore": bson.M{"$gt": 0},
	}

	if communityId != "" {
		filter["communityId"] = communityId
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.controversyScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetCommunityTopPosts(communityId string, sortBy string, limit int) ([]*model.Post, error) {
	filter := bson.M{
		"communityId": communityId,
		"status":      model.PostStatusActive,
	}

	var sortField string
	switch sortBy {
	case "views":
		sortField = "analytics.totalViews"
	case "engagement":
		sortField = "analytics.engagementRate"
	case "quality":
		sortField = "analytics.qualityScore"
	case "trending":
		sortField = "analytics.trendingScore"
	case "hot":
		sortField = "analytics.hotScore"
	default:
		sortField = "analytics.popularityScore"
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: sortField, Value: -1}},
		Limit: int64Ptr(int64(limit)),
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetCommunityTrendingAnalytics(communityId string) (*model.CommunityPostStats, error) {
	filter := bson.M{
		"communityId": communityId,
		"status":      model.PostStatusActive,
	}

	posts, err := s.postQB.Query(s.ctx).FindAll(filter, nil)
	if err != nil {
		return nil, err
	}

	stats := &model.CommunityPostStats{
		CommunityId: communityId,
		TotalPosts:  int64(len(posts)),
	}

	var totalViews, totalEngagement int64
	var totalQuality float64
	postsByType := make(map[string]int64)

	for _, post := range posts {
		if post.Analytics != nil {
			totalViews += post.Analytics.TotalViews
			totalEngagement += post.Analytics.TotalUpvotes + post.Analytics.TotalDownvotes +
				post.Analytics.TotalComments + post.Analytics.TotalShares
			totalQuality += post.Analytics.QualityScore
		}
		postsByType[string(post.Type)]++
	}

	stats.TotalViews = totalViews
	if len(posts) > 0 {
		stats.AvgEngagement = float64(totalEngagement) / float64(len(posts))
		stats.AvgQuality = totalQuality / float64(len(posts))
	}
	stats.PostsByType = postsByType

	return stats, nil
}

func (s *postAnalyticsService) GetCommunityEngagementLeaders(communityId string, limit int) ([]*model.Post, error) {
	filter := bson.M{
		"communityId": communityId,
		"status":      model.PostStatusActive,
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.engagementRate", Value: -1}},
		Limit: int64Ptr(int64(limit)),
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetCommunityRisingStars(communityId string, limit int) ([]*model.Post, error) {
	filter := bson.M{
		"communityId": communityId,
		"status":      model.PostStatusActive,
		"createdAt":   bson.M{"$gte": primitive.NewDateTimeFromTime(time.Now().Add(-24 * time.Hour))},
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.risingScore", Value: -1}},
		Limit: int64Ptr(int64(limit)),
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) GetPostInsights(postId string, days int) (*model.PostInsights, error) {
	post, err := s.GetPostAnalytics(postId)
	if err != nil {
		return nil, err
	}

	if post.Analytics == nil {
		return nil, fmt.Errorf("no analytics data found for post %s", postId)
	}

	insights := &model.PostInsights{
		PostId:             post.PostId,
		Title:              post.Title,
		TotalViews:         post.Analytics.TotalViews,
		UniqueViews:        post.Analytics.UniqueViews,
		EngagementRate:     post.Analytics.EngagementRate,
		ViralityScore:      post.Analytics.ViralityScore,
		QualityScore:       post.Analytics.QualityScore,
		TrendingScore:      post.Analytics.TrendingScore,
		HotScore:           post.Analytics.HotScore,
		RisingScore:        post.Analytics.RisingScore,
		ViewVelocity:       post.Analytics.ViewVelocity1h,
		EngagementVelocity: post.Analytics.EngagementVelocity1h,
		Demographics:       post.Analytics.ViewerDemographics,
		EngagementBreakdown: &model.EngagementBreakdown{
			Upvotes:   post.Analytics.TotalUpvotes,
			Downvotes: post.Analytics.TotalDownvotes,
			Comments:  post.Analytics.TotalComments,
			Shares:    post.Analytics.TotalShares,
			Saves:     post.Analytics.TotalSaves,
			Reports:   post.Analytics.TotalReports,
		},
		MomentumIndicators: &model.MomentumData{
			ViewMomentum:       post.Analytics.ViewMomentum,
			EngagementMomentum: post.Analytics.EngagementMomentum,
			TrendingMomentum:   post.Analytics.TrendingMomentum,
		},
	}

	return insights, nil
}

func (s *postAnalyticsService) GetPostPerformanceComparison(postIds []string) ([]*model.PostComparisonData, error) {
	filter := bson.M{
		"postId": bson.M{"$in": postIds},
	}

	posts, err := s.postQB.Query(s.ctx).FindAll(filter, nil)
	if err != nil {
		return nil, err
	}

	var comparisons []*model.PostComparisonData
	for _, post := range posts {
		comparison := &model.PostComparisonData{
			PostId: post.PostId,
			Title:  post.Title,
			Views:  int64(post.ViewCount),
		}

		if post.Analytics != nil {
			comparison.EngagementRate = post.Analytics.EngagementRate
			comparison.QualityScore = post.Analytics.QualityScore
			comparison.TrendingScore = post.Analytics.TrendingScore
			comparison.HotScore = post.Analytics.HotScore
			comparison.ViewVelocity = post.Analytics.ViewVelocity1h
			comparison.Age = int(post.Analytics.AgeInHours)
		}

		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

func (s *postAnalyticsService) GetAuthorPostAnalytics(authorId string, days int) (*model.AuthorPostStats, error) {
	timeThreshold := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	filter := bson.M{
		"authorId":  authorId,
		"status":    model.PostStatusActive,
		"createdAt": bson.M{"$gte": primitive.NewDateTimeFromTime(timeThreshold)},
	}

	posts, err := s.postQB.Query(s.ctx).FindAll(filter, nil)
	if err != nil {
		return nil, err
	}

	stats := &model.AuthorPostStats{
		AuthorId:   authorId,
		TotalPosts: int64(len(posts)),
	}

	var totalViews, totalEngagement int64
	var totalQuality, totalTrending float64
	communityMap := make(map[string]int)
	var bestPost *model.Post
	var bestScore float64

	for _, post := range posts {
		totalViews += int64(post.ViewCount)
		communityMap[post.CommunityId]++

		if post.Analytics != nil {
			engagement := post.Analytics.TotalUpvotes + post.Analytics.TotalDownvotes +
				post.Analytics.TotalComments + post.Analytics.TotalShares
			totalEngagement += engagement
			totalQuality += post.Analytics.QualityScore
			totalTrending += post.Analytics.TrendingScore

			if post.Analytics.PopularityScore > bestScore {
				bestScore = post.Analytics.PopularityScore
				bestPost = post
			}
		}
	}

	stats.TotalViews = totalViews
	stats.TotalEngagement = totalEngagement
	if len(posts) > 0 {
		stats.AvgQualityScore = totalQuality / float64(len(posts))
		stats.AvgTrendingScore = totalTrending / float64(len(posts))
	}
	stats.BestPerforming = bestPost

	var topCommunities []string
	for community := range communityMap {
		topCommunities = append(topCommunities, community)
		if len(topCommunities) >= 5 {
			break
		}
	}
	stats.TopCommunities = topCommunities

	return stats, nil
}

func (s *postAnalyticsService) SearchPostsByMetrics(criteria model.PostSearchCriteria) ([]*model.Post, error) {
	filter := bson.M{
		"status": model.PostStatusActive,
	}

	if criteria.MinViews > 0 || criteria.MaxViews > 0 {
		viewFilter := bson.M{}
		if criteria.MinViews > 0 {
			viewFilter["$gte"] = criteria.MinViews
		}
		if criteria.MaxViews > 0 {
			viewFilter["$lte"] = criteria.MaxViews
		}
		filter["analytics.totalViews"] = viewFilter
	}

	if criteria.MinEngagement > 0 || criteria.MaxEngagement > 0 {
		engagementFilter := bson.M{}
		if criteria.MinEngagement > 0 {
			engagementFilter["$gte"] = criteria.MinEngagement
		}
		if criteria.MaxEngagement > 0 {
			engagementFilter["$lte"] = criteria.MaxEngagement
		}
		filter["analytics.engagementRate"] = engagementFilter
	}

	if criteria.MinQualityScore > 0 {
		filter["analytics.qualityScore"] = bson.M{"$gte": criteria.MinQualityScore}
	}

	if criteria.MinTrendingScore > 0 {
		filter["analytics.trendingScore"] = bson.M{"$gte": criteria.MinTrendingScore}
	}

	if criteria.MinHotScore > 0 {
		filter["analytics.hotScore"] = bson.M{"$gte": criteria.MinHotScore}
	}

	if criteria.MinViewVelocity > 0 {
		filter["analytics.viewVelocity1h"] = bson.M{"$gte": criteria.MinViewVelocity}
	}

	if len(criteria.PostTypes) > 0 {
		filter["type"] = bson.M{"$in": criteria.PostTypes}
	}

	if len(criteria.Communities) > 0 {
		filter["communityId"] = bson.M{"$in": criteria.Communities}
	}

	if len(criteria.Authors) > 0 {
		filter["authorId"] = bson.M{"$in": criteria.Authors}
	}

	if len(criteria.Tags) > 0 {
		filter["tags"] = bson.M{"$in": criteria.Tags}
	}

	if !criteria.CreatedAfter.Time().IsZero() || !criteria.CreatedBefore.Time().IsZero() {
		dateFilter := bson.M{}
		if !criteria.CreatedAfter.Time().IsZero() {
			dateFilter["$gte"] = criteria.CreatedAfter
		}
		if !criteria.CreatedBefore.Time().IsZero() {
			dateFilter["$lte"] = criteria.CreatedBefore
		}
		filter["createdAt"] = dateFilter
	}

	var sortField string
	var sortOrder int

	switch criteria.SortBy {
	case "views":
		sortField = "analytics.totalViews"
	case "engagement":
		sortField = "analytics.engagementRate"
	case "quality":
		sortField = "analytics.qualityScore"
	case "trending":
		sortField = "analytics.trendingScore"
	case "hot":
		sortField = "analytics.hotScore"
	case "velocity":
		sortField = "analytics.viewVelocity1h"
	case "created":
		sortField = "createdAt"
	default:
		sortField = "analytics.popularityScore"
	}

	if criteria.SortOrder == "asc" {
		sortOrder = 1
	} else {
		sortOrder = -1
	}

	limit := criteria.Limit
	if limit <= 0 {
		limit = 50
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: sortField, Value: sortOrder}},
		Limit: int64Ptr(int64(limit)),
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) CalculateAndUpdatePostScores(postId string) error {
	post, err := s.GetPostAnalytics(postId)
	if err != nil {
		return err
	}

	if post.Analytics == nil {
		post.Analytics = &model.PostAnalytics{}
	}

	now := time.Now()
	createdAt := post.CreatedAt.Time()
	ageInHours := now.Sub(createdAt).Hours()
	post.Analytics.AgeInHours = ageInHours

	agePenalty := 1.0 / (1.0 + ageInHours/24.0)
	post.Analytics.AgePenalty = agePenalty

	freshnessBoost := 1.0
	if ageInHours < 1 {
		freshnessBoost = 2.0
	} else if ageInHours < 6 {
		freshnessBoost = 1.5
	} else if ageInHours < 24 {
		freshnessBoost = 1.2
	}
	post.Analytics.FreshnessBoost = freshnessBoost

	totalEngagement := post.Analytics.TotalUpvotes + post.Analytics.TotalDownvotes +
		post.Analytics.TotalComments + post.Analytics.TotalShares
	if post.Analytics.TotalViews > 0 {
		post.Analytics.EngagementRate = float64(totalEngagement) / float64(post.Analytics.TotalViews)
	}

	if post.Analytics.TotalViews > 0 {
		post.Analytics.CommentToViewRatio = float64(post.Analytics.TotalComments) / float64(post.Analytics.TotalViews)
		post.Analytics.ShareToViewRatio = float64(post.Analytics.TotalShares) / float64(post.Analytics.TotalViews)
		post.Analytics.SaveToViewRatio = float64(post.Analytics.TotalSaves) / float64(post.Analytics.TotalViews)
	}

	totalVotes := post.Analytics.TotalUpvotes + post.Analytics.TotalDownvotes
	if totalVotes > 0 {
		post.Analytics.UpvoteRatio = float64(post.Analytics.TotalUpvotes) / float64(totalVotes)
	}

	if totalVotes > 0 {
		upvoteRatio := post.Analytics.UpvoteRatio
		post.Analytics.ControversyScore = 4 * upvoteRatio * (1 - upvoteRatio)
	}

	if ageInHours > 0 {
		post.Analytics.ViewVelocity1h = float64(post.Analytics.ActivityBuckets.CurrentHour.Views)
		post.Analytics.ViewVelocity6h = float64(post.Analytics.ActivityBuckets.Current6Hours.Views) / 6.0
		post.Analytics.EngagementVelocity1h = float64(
			post.Analytics.ActivityBuckets.CurrentHour.Upvotes +
				post.Analytics.ActivityBuckets.CurrentHour.Downvotes +
				post.Analytics.ActivityBuckets.CurrentHour.Comments +
				post.Analytics.ActivityBuckets.CurrentHour.Shares)
		post.Analytics.EngagementVelocity6h = float64(
			post.Analytics.ActivityBuckets.Current6Hours.Upvotes+
				post.Analytics.ActivityBuckets.Current6Hours.Downvotes+
				post.Analytics.ActivityBuckets.Current6Hours.Comments+
				post.Analytics.ActivityBuckets.Current6Hours.Shares) / 6.0
	}

	currentHourActivity := float64(post.Analytics.ActivityBuckets.CurrentHour.Views)
	previousHourActivity := float64(post.Analytics.ActivityBuckets.PreviousHour.Views)
	post.Analytics.ViewMomentum = currentHourActivity - previousHourActivity

	currentEngagement := post.Analytics.EngagementVelocity1h
	previousEngagement := float64(
		post.Analytics.ActivityBuckets.PreviousHour.Upvotes +
			post.Analytics.ActivityBuckets.PreviousHour.Downvotes +
			post.Analytics.ActivityBuckets.PreviousHour.Comments +
			post.Analytics.ActivityBuckets.PreviousHour.Shares)
	post.Analytics.EngagementMomentum = currentEngagement - previousEngagement

	netVotes := post.Analytics.TotalUpvotes - post.Analytics.TotalDownvotes
	order := math.Log10(math.Max(math.Abs(float64(netVotes)), 1))
	sign := 1.0
	if netVotes < 0 {
		sign = -1.0
	} else if netVotes == 0 {
		sign = 0.0
	}
	seconds := createdAt.Unix() - 1134028003
	post.Analytics.HotScore = sign*order + float64(seconds)/45000.0

	viewWeight := math.Log10(math.Max(float64(post.Analytics.TotalViews), 1))
	engagementWeight := math.Log10(math.Max(float64(totalEngagement), 1)) * 2
	timeDecay := math.Exp(-ageInHours / 24.0)
	post.Analytics.TrendingScore = (viewWeight + engagementWeight) * timeDecay * freshnessBoost

	qualityFactors := []float64{
		post.Analytics.CommentToViewRatio * 100,
		post.Analytics.ShareToViewRatio * 200,
		post.Analytics.SaveToViewRatio * 150,
		post.Analytics.UpvoteRatio * 100,
		math.Min(post.Analytics.EngagementRate*100, 50),
	}

	qualitySum := 0.0
	for _, factor := range qualityFactors {
		qualitySum += factor
	}
	post.Analytics.QualityScore = qualitySum / float64(len(qualityFactors))

	shareBoost := math.Log10(math.Max(float64(post.Analytics.TotalShares), 1)) * 20
	velocityBoost := post.Analytics.ViewVelocity1h * 2
	momentumBoost := math.Max(post.Analytics.ViewMomentum, 0) * 5
	post.Analytics.ViralityScore = shareBoost + velocityBoost + momentumBoost

	viewScore := math.Log10(math.Max(float64(post.Analytics.TotalViews), 1)) * 10
	engagementScore := math.Log10(math.Max(float64(totalEngagement), 1)) * 15
	qualityBonus := post.Analytics.QualityScore * 0.5
	post.Analytics.PopularityScore = viewScore + engagementScore + qualityBonus

	if ageInHours < 48 {
		risingFactor := (48 - ageInHours) / 48.0
		momentumFactor := math.Max(post.Analytics.EngagementMomentum, 0) * 10
		velocityFactor := post.Analytics.EngagementVelocity1h * 5
		post.Analytics.RisingScore = risingFactor * (momentumFactor + velocityFactor)
	}

	post.Analytics.TrendingMomentum = post.Analytics.ViewMomentum + post.Analytics.EngagementMomentum

	post.Analytics.LastScoreUpdateAt = primitive.NewDateTimeFromTime(now)

	filter := bson.M{"postId": postId}
	update := bson.M{
		"$set": bson.M{
			"analytics": post.Analytics,
			"updatedAt": primitive.NewDateTimeFromTime(now),
		},
	}

	_, err = s.postQB.Query(s.ctx).UpdateOne(filter, update, nil)
	return err
}

func (s *postAnalyticsService) GetPostsRequiringScoreUpdate() ([]*model.Post, error) {

	oneHourAgo := time.Now().Add(-1 * time.Hour)

	filter := bson.M{
		"status": model.PostStatusActive,
		"$or": []bson.M{
			{"analytics.lastScoreUpdateAt": bson.M{"$lt": primitive.NewDateTimeFromTime(oneHourAgo)}},
			{"analytics.lastScoreUpdateAt": bson.M{"$exists": false}},
		},
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "lastActivityAt", Value: -1}},
		Limit: int64Ptr(100),
	}

	return s.postQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *postAnalyticsService) CleanupOldPostAnalytics(olderThanDays int) error {
	cutoffTime := time.Now().Add(-time.Duration(olderThanDays) * 24 * time.Hour)

	filter := bson.M{
		"createdAt": bson.M{"$lt": primitive.NewDateTimeFromTime(cutoffTime)},
	}

	update := bson.M{
		"$unset": bson.M{
			"analytics.uniqueViewers":        "",
			"analytics.uniqueEngagers":       "",
			"analytics.viewerDemographics":   "",
			"analytics.activityBuckets":      "",
			"analytics.weightedViews1h":      "",
			"analytics.weightedViews6h":      "",
			"analytics.weightedEngagement1h": "",
			"analytics.weightedEngagement6h": "",
		},
	}

	_, err := s.postQB.Query(s.ctx).UpdateMany(filter, update, nil)
	return err
}

func (s *postAnalyticsService) ArchivePostAnalytics(postIds []string) error {
	filter := bson.M{
		"postId": bson.M{"$in": postIds},
	}

	update := bson.M{
		"$set": bson.M{
			"analytics.archived": true,
			"updatedAt":          primitive.NewDateTimeFromTime(time.Now()),
		},
		"$unset": bson.M{
			"analytics.uniqueViewers":      "",
			"analytics.uniqueEngagers":     "",
			"analytics.viewerDemographics": "",
			"analytics.activityBuckets":    "",
		},
	}

	_, err := s.postQB.Query(s.ctx).UpdateMany(filter, update, nil)
	return err
}

func int64Ptr(v int64) *int64 {
	return &v
}
