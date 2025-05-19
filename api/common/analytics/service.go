package analytics

import (
	community "sync-backend/api/community/model"
	post "sync-backend/api/post/model"
	"sync-backend/arch/mongo"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnalyticsService interface {
	ApplyGetPost(postId string) error
	ApplyJoinCommunity(communityId string) error
	ApplyLeaveCommunity(communityId string) error
}

type analyticsService struct {
	transactionBuilder mongo.TransactionBuilder
}

func NewAnalyticsService(db mongo.Database) AnalyticsService {
	return &analyticsService{
		transactionBuilder: mongo.NewTransactionBuilder(db),
	}
}
func (s *analyticsService) ApplyGetPost(postId string) error {
	// Start a transaction
	tx := s.transactionBuilder.GetTransaction(mongo.DefaultShortTransactionTimeout)

	if err := tx.Start(); err != nil {
		return err
	}

	err := tx.PerformTransaction(func(session mongo.DatabaseSession) error {
		postCollection := session.Collection(post.PostCollectionName)
		communityCollection := session.Collection(community.CommunityCollectionName)

		// Update the post and get its data in one operation
		var postModel post.Post
		err := postCollection.FindOneAndUpdate(
			bson.M{"postId": postId, "status": post.PostStatusActive},
			bson.M{
				"$inc": bson.M{"viewCount": 1},
				"$set": bson.M{"lastActivity": primitive.NewDateTimeFromTime(time.Now())},
			},
		).Decode(&postModel)

		if err != nil {
			return err
		}

		// Update the community stats
		_, err = communityCollection.UpdateOne(
			bson.M{"communityId": postModel.CommunityId},
			bson.M{
				"$inc": bson.M{
					"stats.dailyActiveUsers":   1,
					"stats.weeklyActiveUsers":  1,
					"stats.monthlyActiveUsers": 1,
				},
			},
		)
		return err
	})
	return err
}

func (s *analyticsService) ApplyJoinCommunity(communityId string) error {
	// Start a transaction
	tx := s.transactionBuilder.GetTransaction(mongo.DefaultShortTransactionTimeout)
	if err := tx.Start(); err != nil {
		return err
	}

	err := tx.PerformTransaction(func(session mongo.DatabaseSession) error {
		communityCollection := session.Collection(community.CommunityCollectionName)

		// Update the community stats
		var communityData community.Community
		err := communityCollection.FindOne(
			bson.M{"communityId": communityId},
		).Decode(&communityData)
		if err != nil {
			return err
		}

		newGrowthRate := 0.7 * communityData.Stats.GrowthRate
		if communityData.MemberCount > 0 {
			newGrowthRate += 0.3 * (1.0 / float64(communityData.MemberCount))
		} else {
			newGrowthRate = 1.0
		}

		_, err = communityCollection.UpdateOne(
			bson.M{"communityId": communityId},
			bson.M{
				"$inc": bson.M{
					"stats.dailyActiveUsers":   1,
					"stats.weeklyActiveUsers":  1,
					"stats.monthlyActiveUsers": 1,
					"stats.memberCount":        1,
				},
				"$set": bson.M{
					"stats.growthRate": newGrowthRate,
					"lastActivity":     primitive.NewDateTimeFromTime(time.Now()),
				},
			},
		)
		return err
	})

	return err
}
func (s *analyticsService) ApplyLeaveCommunity(communityId string) error {
	// Start a transaction
	tx := s.transactionBuilder.GetTransaction(mongo.DefaultShortTransactionTimeout)

	if err := tx.Start(); err != nil {
		return err
	}

	err := tx.PerformTransaction(func(session mongo.DatabaseSession) error {
		communityCollection := session.Collection(community.CommunityCollectionName)
		// Get current community data for calculating growth rate
		var communityData community.Community
		err := communityCollection.FindOne(
			bson.M{"communityId": communityId},
		).Decode(&communityData)
		if err != nil {
			return err
		}

		newGrowthRate := 0.7 * communityData.Stats.GrowthRate
		if communityData.MemberCount > 1 { // Check if there are still members left
			newGrowthRate -= 0.3 * (1.0 / float64(communityData.MemberCount))
		} else {
			newGrowthRate = 0.0 // No growth if no members left
		}

		// Update the community stats
		_, err = communityCollection.UpdateOne(
			bson.M{"communityId": communityId},
			bson.M{
				"$inc": bson.M{
					"stats.dailyActiveUsers":   -1,
					"stats.weeklyActiveUsers":  -1,
					"stats.monthlyActiveUsers": -1,
					"stats.memberCount":        -1,
				},
				"$set": bson.M{
					"stats.growthRate": newGrowthRate,
					"lastActivity":     primitive.NewDateTimeFromTime(time.Now()),
				},
			},
		)
		return err
	})
	return err
}
