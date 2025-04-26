package location

import (
	pg "sync-backend/arch/postgres"

	"sync-backend/api/location/model"
	"sync-backend/utils"
)

type LocationService interface {
	GetLocationByIp(ip string, localCode string) (*model.UserLocationInfo, error)
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

func (s *locationService) GetLocationByIp(ip string, localCode string) (*model.UserLocationInfo, error) {
	query := `SELECT country_name, city_name, latitude, longitude, accuracy_radius 
		FROM geoip2_network net
		LEFT JOIN geoip2_location location ON (
			net.geoname_id = location.geoname_id
			AND location.locale_code = $1
		)
		WHERE network >>= $2`

	locationData, err := s.ipQueryBuilder.SingleQuery().FilterOne(query, localCode, ip)
	if err != nil {
		s.log.Error("Error getting location by IP: %s, localCode: %s, error: %v", ip, localCode, err)
		return nil, err
	}
	locationData.Accuracy = locationData.Accuracy + " km"

	return locationData, nil
}
