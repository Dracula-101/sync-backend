package analytics

import (
	"sync-backend/arch/mongo"
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
	return nil
}

func (s *analyticsService) ApplyJoinCommunity(communityId string) error {
	return nil
}

func (s *analyticsService) ApplyLeaveCommunity(communityId string) error {
	return nil
}
