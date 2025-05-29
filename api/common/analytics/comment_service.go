package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"sync-backend/api/comment/model"
	postModel "sync-backend/api/post/model"
	"sync-backend/arch/mongo"
)

type CommentAnalytics interface {
	RecordCommentView(commentId, userId string) error
	RecordCommentVote(commentId, userId string, voteType postModel.VoteType) error
	RecordCommentReply(commentId, userId string) error
	RecordCommentReaction(commentId, userId string, reactionType model.ReactionType) error
	RecordCommentReport(commentId, userId string) error

	GetCommentAnalytics(commentId string) (*model.Comment, error)
	GetTopCommentsByEngagement(postId string, limit int) ([]*model.Comment, error)
	GetHotComments(postId string, limit int) ([]*model.Comment, error)
	GetControversialComments(postId string, limit int) ([]*model.Comment, error)
	GetTopCommentsByQuality(postId string, limit int) ([]*model.Comment, error)

	GetPostCommentsAnalytics(postId string) (*model.PostCommentStats, error)
	GetAuthorCommentsAnalytics(authorId string, days int) ([]*model.Comment, error)
	GetCommentInsights(commentId string) (*model.CommentInsights, error)
	GetCommentThread(parentCommentId string, limit int) ([]*model.Comment, error)

	CalculateAndUpdateCommentScores(commentId string) error
	GetCommentsRequiringScoreUpdate() ([]*model.Comment, error)
}

type commentAnalyticsService struct {
	commentQB mongo.QueryBuilder[model.Comment]
	ctx       context.Context
}

func NewCommentAnalyticsService(db mongo.Database) CommentAnalytics {
	return &commentAnalyticsService{
		commentQB: mongo.NewQueryBuilder[model.Comment](db, model.CommentCollectionName),
		ctx:       context.Background(),
	}
}

func (s *commentAnalyticsService) RecordCommentView(commentId, userId string) error {
	filter := bson.M{"commentId": commentId}

	update := bson.M{
		"$inc": bson.M{
			"analytics.totalViews": 1,
		},
		"$set": bson.M{
			"status":                   model.CommentStatusActive,
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	// Set the first view timestamp if it doesn't exist yet
	if userId != "" {
		update["$addToSet"] = bson.M{
			"uniqueViewers": userId,
		}
		update["$inc"].(bson.M)["analytics.uniqueViews"] = 1
	}

	// Set the first view timestamp if it doesn't exist
	update["$setOnInsert"] = bson.M{
		"analytics.firstViewAt": primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err := s.commentQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdateCommentScores(commentId)
}

func (s *commentAnalyticsService) RecordCommentVote(commentId, userId string, voteType postModel.VoteType) error {
	filter := bson.M{"commentId": commentId}

	// Initialize the update operation
	update := bson.M{
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	// Apply the vote based on type
	if voteType == postModel.Upvote {
		// Upvote (increase synergy)
		update["$inc"] = bson.M{
			"analytics.totalUpvotes": 1,
		}
	} else if voteType == postModel.Downvote {
		// Downvote (decrease synergy)
		update["$inc"] = bson.M{
			"analytics.totalDownvotes": 1,
		}
	}

	_, err := s.commentQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdateCommentScores(commentId)
}

func (s *commentAnalyticsService) RecordCommentReply(commentId, userId string) error {
	filter := bson.M{"commentId": commentId}

	update := bson.M{
		"$inc": bson.M{
			"analytics.totalReplies": 1,
		},
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := s.commentQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdateCommentScores(commentId)
}

func (s *commentAnalyticsService) RecordCommentReaction(commentId, userId string, reactionType model.ReactionType) error {
	filter := bson.M{"commentId": commentId}

	update := bson.M{
		"$inc": bson.M{
			"analytics.totalReactions":                          1,
			"analytics.reactionsByType." + string(reactionType): 1,
		},
		"$set": bson.M{
			"analytics.lastActivityAt": primitive.NewDateTimeFromTime(time.Now()),
		},
		"$addToSet": bson.M{
			"reactions": model.Reaction{
				UserId:    userId,
				Type:      reactionType,
				CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
			},
		},
	}

	_, err := s.commentQB.Query(s.ctx).UpdateOne(filter, update, nil)
	if err != nil {
		return err
	}

	return s.CalculateAndUpdateCommentScores(commentId)
}

func (s *commentAnalyticsService) RecordCommentReport(commentId, userId string) error {
	filter := bson.M{"commentId": commentId}

	update := bson.M{
		"$inc": bson.M{
			"moderationInfo.reportCount": 1,
		},
		"$set": bson.M{
			"moderationInfo.lastReportedAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := s.commentQB.Query(s.ctx).UpdateOne(filter, update, nil)
	return err
}

func (s *commentAnalyticsService) GetCommentAnalytics(commentId string) (*model.Comment, error) {
	filter := bson.M{"commentId": commentId}
	return s.commentQB.Query(s.ctx).FindOne(filter, nil)
}

func (s *commentAnalyticsService) GetTopCommentsByEngagement(postId string, limit int) ([]*model.Comment, error) {
	filter := bson.M{
		"postId": postId,
		"status": model.CommentStatusActive,
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.engagementRate", Value: -1}},
		Limit: &limitVal,
	}

	return s.commentQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *commentAnalyticsService) GetHotComments(postId string, limit int) ([]*model.Comment, error) {
	filter := bson.M{
		"postId": postId,
		"status": model.CommentStatusActive,
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.hotScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.commentQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *commentAnalyticsService) GetControversialComments(postId string, limit int) ([]*model.Comment, error) {
	filter := bson.M{
		"postId":                     postId,
		"status":                     model.CommentStatusActive,
		"analytics.controversyScore": bson.M{"$gt": 0},
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.controversyScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.commentQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *commentAnalyticsService) GetTopCommentsByQuality(postId string, limit int) ([]*model.Comment, error) {
	filter := bson.M{
		"postId":                 postId,
		"status":                 model.CommentStatusActive,
		"analytics.qualityScore": bson.M{"$gt": 0},
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.qualityScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.commentQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *commentAnalyticsService) GetPostCommentsAnalytics(postId string) (*model.PostCommentStats, error) {
	filter := bson.M{
		"postId": postId,
		"status": model.CommentStatusActive,
	}

	comments, err := s.commentQB.Query(s.ctx).FindAll(filter, nil)
	if err != nil {
		return nil, err
	}

	stats := &model.PostCommentStats{
		PostId:        postId,
		TotalComments: int64(len(comments)),
	}

	var totalViews int64
	var totalEngagement, totalQuality float64
	commentsByLevel := make(map[string]int64)

	for _, comment := range comments {
		if comment.Analytics != nil {
			totalViews += int64(comment.Analytics.TotalViews)
			engagementSum := float64(comment.Analytics.TotalUpvotes + comment.Analytics.TotalDownvotes + comment.Analytics.TotalReplies + comment.Analytics.TotalReactions)
			totalEngagement += engagementSum
			totalQuality += comment.Analytics.QualityScore
		}

		level := fmt.Sprintf("%d", comment.Level)
		commentsByLevel[level]++
	}

	stats.TotalViews = totalViews
	if len(comments) > 0 {
		stats.AvgEngagement = totalEngagement / float64(len(comments))
		stats.AvgQuality = totalQuality / float64(len(comments))
	}
	stats.CommentsByLevel = commentsByLevel

	// Find top comments by quality score
	if len(comments) > 0 {
		topComments, _ := s.GetTopCommentsByQuality(postId, 5)
		stats.TopComments = topComments

		controversialComments, _ := s.GetControversialComments(postId, 5)
		stats.ControversialComments = controversialComments
	}

	return stats, nil
}

func (s *commentAnalyticsService) GetAuthorCommentsAnalytics(authorId string, days int) ([]*model.Comment, error) {
	timeThreshold := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	filter := bson.M{
		"authorId":  authorId,
		"status":    model.CommentStatusActive,
		"createdAt": bson.M{"$gte": primitive.NewDateTimeFromTime(timeThreshold)},
	}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "analytics.engagementRate", Value: -1}},
	}

	return s.commentQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *commentAnalyticsService) GetCommentInsights(commentId string) (*model.CommentInsights, error) {
	comment, err := s.GetCommentAnalytics(commentId)
	if err != nil {
		return nil, err
	}

	if comment.Analytics == nil {
		return nil, fmt.Errorf("no analytics data available for comment %s", commentId)
	}

	// Convert reactionsByType to string keys for JSON
	topReactions := make(map[string]int32)
	for reactionType, count := range comment.Analytics.ReactionsByType {
		topReactions[string(reactionType)] = count
	}

	insights := &model.CommentInsights{
		CommentId:      comment.CommentId,
		PostId:         comment.PostId,
		TotalViews:     comment.Analytics.TotalViews,
		EngagementRate: comment.Analytics.EngagementRate,
		QualityScore:   comment.Analytics.QualityScore,
		HotScore:       comment.Analytics.HotScore,
		TopReactions:   topReactions,
		ResponseCount:  int32(comment.ReplyCount),
	}

	return insights, nil
}

func (s *commentAnalyticsService) GetCommentThread(parentCommentId string, limit int) ([]*model.Comment, error) {
	filter := bson.M{
		"parentId": parentCommentId,
		"status":   model.CommentStatusActive,
	}

	limitVal := int64(limit)
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "analytics.hotScore", Value: -1}},
		Limit: &limitVal,
	}

	return s.commentQB.Query(s.ctx).FindAll(filter, opts)
}

func (s *commentAnalyticsService) CalculateAndUpdateCommentScores(commentId string) error {
	comment, err := s.GetCommentAnalytics(commentId)
	if err != nil {
		return err
	}

	if comment.Analytics == nil {
		return fmt.Errorf("no analytics data available for comment %s", commentId)
	}

	now := time.Now()
	createdAt := comment.CreatedAt.Time()

	// Calculate upvote ratio
	totalVotes := comment.Analytics.TotalUpvotes + comment.Analytics.TotalDownvotes
	if totalVotes > 0 {
		comment.Analytics.UpvoteRatio = float64(comment.Analytics.TotalUpvotes) / float64(totalVotes)
	} else {
		comment.Analytics.UpvoteRatio = 0
	}

	// Calculate controversy score - high when votes are even split
	if totalVotes >= 10 {
		balance := float64(comment.Analytics.TotalUpvotes) / float64(totalVotes)
		comment.Analytics.ControversyScore = 4.0 * balance * (1.0 - balance)
	} else {
		comment.Analytics.ControversyScore = 0
	}

	// Calculate engagement rate
	totalEngagement := comment.Analytics.TotalUpvotes + comment.Analytics.TotalDownvotes +
		comment.Analytics.TotalReplies + comment.Analytics.TotalReactions
	if comment.Analytics.TotalViews > 0 {
		comment.Analytics.EngagementRate = float64(totalEngagement) / float64(comment.Analytics.TotalViews)
	} else {
		comment.Analytics.EngagementRate = 0
	}

	// Calculate hot score (Reddit-like algorithm)
	netVotes := comment.Analytics.TotalUpvotes - comment.Analytics.TotalDownvotes
	order := math.Log10(math.Max(math.Abs(float64(netVotes)), 1))
	sign := 1.0
	if netVotes < 0 {
		sign = -1.0
	} else if netVotes == 0 {
		sign = 0.0
	}
	seconds := createdAt.Unix() - 1134028003
	comment.Analytics.HotScore = sign*order + float64(seconds)/45000.0

	// Calculate quality score
	ageInHours := now.Sub(createdAt).Hours()
	agePenalty := 1.0
	if ageInHours > 24 {
		agePenalty = 24.0 / ageInHours
	}

	upvoteWeight := float64(comment.Analytics.TotalUpvotes) * 1.0
	replyWeight := float64(comment.Analytics.TotalReplies) * 1.5
	reactionWeight := float64(comment.Analytics.TotalReactions) * 0.8
	viewWeight := math.Log10(math.Max(float64(comment.Analytics.TotalViews), 1)) * 0.5

	// Penalize for high downvotes
	downvotePenalty := math.Min(0.8, float64(comment.Analytics.TotalDownvotes)/float64(math.Max(float64(comment.Analytics.TotalUpvotes), 1))*0.8)

	comment.Analytics.QualityScore = (upvoteWeight + replyWeight + reactionWeight + viewWeight) * agePenalty * (1.0 - downvotePenalty)

	// Update the comment with calculated scores
	update := bson.M{
		"$set": bson.M{
			"analytics": comment.Analytics,
		},
	}

	_, err = s.commentQB.Query(s.ctx).UpdateOne(bson.M{"commentId": commentId}, update, nil)
	return err
}

func (s *commentAnalyticsService) GetCommentsRequiringScoreUpdate() ([]*model.Comment, error) {
	// Get comments that have been active in the last hour but haven't had scores updated
	oneHourAgo := primitive.NewDateTimeFromTime(time.Now().Add(-1 * time.Hour))
	filter := bson.M{
		"analytics.lastActivityAt": bson.M{"$gt": oneHourAgo},
	}

	return s.commentQB.Query(s.ctx).FindAll(filter, nil)
}
