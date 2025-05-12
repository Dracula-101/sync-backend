package location

import (
	pg "sync-backend/arch/postgres"

	"sync-backend/api/common/location/model"
	"sync-backend/utils"
)

type LocationService interface {
	GetLocationByIp(ip string) (*model.UserLocationInfo, error)
	GetLocationByLocaleCode(localeCode string) (*model.UserLocationInfo, error)
}

type locationService struct {
	log            utils.AppLogger
	ipQueryBuilder pg.QueryBuilder[model.UserLocationInfo]
}

func NewLocationService(db pg.Database) LocationService {
	return &locationService{
		log:            utils.NewServiceLogger("LocationService"),
		ipQueryBuilder: pg.NewQueryBuilder[model.UserLocationInfo](db),
	}
}

func (s *locationService) GetLocationByIp(ip string) (*model.UserLocationInfo, error) {
	query := `SELECT country_name, city_name, latitude, longitude, time_zone, gmt_offset, locale_code 
		FROM geoip2_network net
		LEFT JOIN geoip2_location location ON (
			net.geoname_id = location.geoname_id
		)
		WHERE network >>= $1`

	locationData, err := s.ipQueryBuilder.SingleQuery().FilterOne(query, ip)
	if err != nil {
		s.log.Error("Error getting location by IP: %s, error: %v", ip, err)
		return nil, err
	}
	if locationData == nil {
		return model.NewUserLocationInfo(
			"Unknown Country",
			"Unknown City",
			0,
			0,
			"Unknown Timezone",
			"Unknown GMT",
			"Unknown Local",
		), nil
	}
	return locationData, nil
}

func (s *locationService) GetLocationByLocaleCode(localeCode string) (*model.UserLocationInfo, error) {
	query := `SELECT country_name, city_name, latitude, longitude, time_zone, gmt_offset, locale_code 
		FROM geoip2_network net
		LEFT JOIN geoip2_location location ON (
			net.geoname_id = location.geoname_id
		)
		WHERE locale_code = $1`

	locationData, err := s.ipQueryBuilder.SingleQuery().FilterOne(query, localeCode)
	if err != nil {
		s.log.Error("Error getting location by locale code: %s, error: %v", localeCode, err)
		return nil, err
	}
	if locationData == nil {
		return model.NewUserLocationInfo(
			"Unknown Country",
			"Unknown City",
			0,
			0,
			"Unknown Timezone",
			"Unknown GMT",
			"Unknown Local",
		), nil
	}
	return locationData, nil
}
