package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommunityAnalytics struct {
	LastActivityAt     primitive.DateTime `bson:"lastActivityAt" json:"lastActivityAt"`
	LastPostAt         primitive.DateTime `bson:"lastPostAt" json:"lastPostAt"`
	LastCommentAt      primitive.DateTime `bson:"lastCommentAt" json:"lastCommentAt"`
	ActivityBuckets    ActivityBuckets    `bson:"activityBuckets" json:"activityBuckets"`
	TotalViews         int64              `bson:"totalViews" json:"totalViews"`
	TotalLikes         int64              `bson:"totalLikes" json:"totalLikes"`
	TotalComments      int64              `bson:"totalComments" json:"totalComments"`
	TotalShares        int64              `bson:"totalShares" json:"totalShares"`
	ActiveMembersToday int64              `bson:"activeMembersToday" json:"activeMembersToday"`
	ActiveMembersWeek  int64              `bson:"activeMembersWeek" json:"activeMembersWeek"`
	ActiveMembersMonth int64              `bson:"activeMembersMonth" json:"activeMembersMonth"`
	MemberJoinsToday   int64              `bson:"memberJoinsToday" json:"memberJoinsToday"`
	MemberJoinsWeek    int64              `bson:"memberJoinsWeek" json:"memberJoinsWeek"`
	MemberJoinsMonth   int64              `bson:"memberJoinsMonth" json:"memberJoinsMonth"`
	EngagementScore    float64            `bson:"engagementScore" json:"engagementScore"`
	TrendingScore      float64            `bson:"trendingScore" json:"trendingScore"`
	QualityScore       float64            `bson:"qualityScore" json:"qualityScore"`
	ScoresUpdatedAt    primitive.DateTime `bson:"scoresUpdatedAt" json:"scoresUpdatedAt"`
}

// ActivityBuckets stores activity counts in time buckets for trending calculations
type ActivityBuckets struct {
	CurrentHour ActivityBucket   `bson:"currentHour" json:"currentHour"`
	Last24Hours []ActivityBucket `bson:"last24Hours" json:"last24Hours"`
	Last7Days   []ActivityBucket `bson:"last7Days" json:"last7Days"`
	Last30Days  []ActivityBucket `bson:"last30Days" json:"last30Days"`
}

type ActivityBucket struct {
	Timestamp   primitive.DateTime `bson:"timestamp" json:"timestamp"`
	Posts       int64              `bson:"posts" json:"posts"`
	Comments    int64              `bson:"comments" json:"comments"`
	Likes       int64              `bson:"likes" json:"likes"`
	Views       int64              `bson:"views" json:"views"`
	Shares      int64              `bson:"shares" json:"shares"`
	NewMembers  int64              `bson:"newMembers" json:"newMembers"`
	ActiveUsers int64              `bson:"activeUsers" json:"activeUsers"`
}

func NewActivityBuckets() ActivityBuckets {
	now := primitive.NewDateTimeFromTime(time.Now())
	return ActivityBuckets{
		CurrentHour: ActivityBucket{
			Timestamp:   now,
			Posts:       0,
			Comments:    0,
			Likes:       0,
			Views:       0,
			Shares:      0,
			NewMembers:  1,
			ActiveUsers: 1,
		},
		Last24Hours: []ActivityBucket{},
		Last7Days:   []ActivityBucket{},
		Last30Days:  []ActivityBucket{},
	}
}

func NewCommunityAnalytics() *CommunityAnalytics {
	return &CommunityAnalytics{
		LastActivityAt:     primitive.NewDateTimeFromTime(time.Now()),
		ActivityBuckets:    NewActivityBuckets(),
		TotalViews:         0,
		TotalLikes:         0,
		TotalComments:      0,
		TotalShares:        0,
		ActiveMembersToday: 0,
		ActiveMembersWeek:  0,
		ActiveMembersMonth: 0,
		MemberJoinsToday:   0,
		MemberJoinsWeek:    0,
		MemberJoinsMonth:   0,
		EngagementScore:    0.0,
		TrendingScore:      0.0,
		QualityScore:       0.0,
	}
}
