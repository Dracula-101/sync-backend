package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// CommentAnalytics contains essential analytics data for a comment
type CommentAnalytics struct {
	TotalViews       int32                  `bson:"totalViews" json:"totalViews"`
	UniqueViews      int32                  `bson:"uniqueViews" json:"uniqueViews"`
	TotalUpvotes     int32                  `bson:"totalUpvotes" json:"totalUpvotes"`
	TotalDownvotes   int32                  `bson:"totalDownvotes" json:"totalDownvotes"`
	TotalReplies     int32                  `bson:"totalReplies" json:"totalReplies"`
	TotalReactions   int32                  `bson:"totalReactions" json:"totalReactions"`
	UpvoteRatio      float64                `bson:"upvoteRatio" json:"upvoteRatio"`           // Upvotes / (upvotes + downvotes)
	EngagementRate   float64                `bson:"engagementRate" json:"engagementRate"`     // Engagement / views
	ControversyScore float64                `bson:"controversyScore" json:"controversyScore"` // Measure of up/downvote balance
	HotScore         float64                `bson:"hotScore" json:"hotScore"`                 // Hot algorithm score (short-term popularity)
	QualityScore     float64                `bson:"qualityScore" json:"qualityScore"`         // Content quality score
	FirstViewAt      primitive.DateTime     `bson:"firstViewAt,omitempty" json:"firstViewAt,omitempty"`
	LastActivityAt   primitive.DateTime     `bson:"lastActivityAt,omitempty" json:"lastActivityAt,omitempty"`
	ReactionsByType  map[ReactionType]int32 `bson:"reactionsByType,omitempty" json:"reactionsByType,omitempty"`
}

// CommentActivityWindow represents a simplified activity window for a comment
type CommentActivityWindow struct {
	Views       int32              `bson:"views" json:"views"`
	Upvotes     int32              `bson:"upvotes" json:"upvotes"`
	Downvotes   int32              `bson:"downvotes" json:"downvotes"`
	Replies     int32              `bson:"replies" json:"replies"`
	Reactions   int32              `bson:"reactions" json:"reactions"`
	WindowStart primitive.DateTime `bson:"windowStart" json:"windowStart"`
}

// CommentInsights contains simplified analytics insights for a comment
type CommentInsights struct {
	CommentId      string           `json:"commentId"`
	PostId         string           `json:"postId"`
	TotalViews     int32            `json:"totalViews"`
	EngagementRate float64          `json:"engagementRate"`
	QualityScore   float64          `json:"qualityScore"`
	HotScore       float64          `json:"hotScore"`
	TopReactions   map[string]int32 `json:"topReactions"`
	ResponseCount  int32            `json:"responseCount"` // Number of direct responses
}

// CommentAnalyticsUpdate represents a minimal structure for updating comment analytics
type CommentAnalyticsUpdate struct {
	Type         string `json:"type"` // view, upvote, downvote, reply, reaction
	UserId       string `json:"userId"`
	CommentId    string `json:"commentId"`
	ReactionType string `json:"reactionType,omitempty"`
}

// PostCommentStats contains analytics for all comments in a post
type PostCommentStats struct {
	PostId                string           `json:"postId"`
	TotalComments         int64            `json:"totalComments"`
	TotalViews            int64            `json:"totalViews"`
	AvgEngagement         float64          `json:"avgEngagement"`
	AvgQuality            float64          `json:"avgQuality"`
	TopComments           []*Comment       `json:"topComments"`
	ControversialComments []*Comment       `json:"controversialComments"`
	CommentsByLevel       map[string]int64 `json:"commentsByLevel"`
	ActivityTrend         string           `json:"activityTrend"`
}
