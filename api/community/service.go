package community

import (
	"sync-backend/arch/network"
	"sync-backend/utils"
)

type CommunityService interface {
}

type communityService struct {
	network.BaseService
	logger utils.AppLogger
}

func NewCommunityService() CommunityService {
	return &communityService{
		BaseService: network.NewBaseService(),
		logger:      utils.NewServiceLogger("CommunityService"),
	}
}
