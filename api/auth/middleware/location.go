package middleware

import (
	"sync-backend/api/common/location"
	locationModel "sync-backend/api/common/location/model"
	"sync-backend/arch/common"
	coredto "sync-backend/arch/dto"
	"sync-backend/arch/network"
	"sync-backend/arch/redis"
	"sync-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type locationProvider struct {
	network.ResponseSender
	common.ContextPayload
	logger          utils.AppLogger
	locationService location.LocationService
	cacheStore      redis.Store
}

func NewLocationProvider(
	locationService location.LocationService,
	cacheStore redis.Store,
) *locationProvider {
	return &locationProvider{
		ResponseSender:  network.NewResponseSender(),
		ContextPayload:  common.NewContextPayload(),
		logger:          utils.NewServiceLogger("AuthProvider"),
		locationService: locationService,
		cacheStore:      cacheStore,
	}
}

func (p *locationProvider) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var locationData *locationModel.UserLocationInfo
		ip := ctx.ClientIP()
		if ip == "" {
			ip = "127.0.0.1"
		}

		cacheKey := "location:" + ip
		storedLocation, err := p.cacheStore.GetInstance().Get(ctx, cacheKey).Result()
		if err != nil {
			p.logger.Debug("Cache miss or error for IP %s: %v", ip, err)
			storedLocation = ""
		}

		if storedLocation == "" {
			// data not found in cache, fetch from service
			locationData, err = p.locationService.GetLocationByIp(ip)
			if err != nil {
				p.logger.Error("Error getting location by IP: %s, error: %v", ip, err)
				p.Send(ctx).MixedError(err)
				return
			}
			if locationData == nil {
				locationData = locationModel.NewUserLocationInfo(
					"Unknown Country",
					"Unknown City",
					0,
					0,
					"Unknown Timezone",
					"Unknown GMT",
					"Unknown Local",
				)
			}

			// Marshal struct and store in cache
			jsonData, err := locationData.MarshalJSON()
			if err != nil {
				p.logger.Warn("Error marshalling location data: %v", err)
			} else {
				err = p.cacheStore.GetInstance().Set(ctx, cacheKey, jsonData, time.Minute*5).Err()
				if err != nil {
					p.logger.Warn("Error storing location data in cache: %v", err)
				}
			}
		} else {
			// data found in cache, unmarshal to struct
			locationData = &locationModel.UserLocationInfo{}
			err = locationData.UnmarshalJSON([]byte(storedLocation))
			if err != nil {
				p.logger.Error("Error unmarshalling cached location data: %v", err)
				p.Send(ctx).MixedError(err)
				return
			}
		}

		savedLocation := &coredto.BaseLocationRequest{
			Country:   locationData.Country,
			City:      locationData.City,
			Latitude:  locationData.Lat,
			Longitude: locationData.Lon,
			TimeZone:  locationData.Timezone,
			GMTOffset: locationData.GmtOffset,
			Locale:    locationData.Locale,
			IpAddress: ip,
		}
		ctx.Set(network.UserLocation, savedLocation)
	}
}
