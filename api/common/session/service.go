package session

import (
	"time"

	"sync-backend/api/common/session/model"
	"sync-backend/arch/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SessionService interface {
	CreateSession(userID string, token string, refreshToken string, expiresAt time.Time, deviceInfo model.DeviceInfo, userLocationInfo model.LocationInfo) (*model.Session, error)
	GetSessionByToken(token string) (*model.Session, error)
	GetSessionByRefreshToken(refreshToken string) (*model.Session, error)
	UpdateSession(sessionID string, accessToken string, refreshToken string, expiresAt time.Time) (*model.Session, error)
	UpdateSessionInfo(sessionID string, deviceInfo model.DeviceInfo, userLocationInfo model.LocationInfo) error
	GetUserActiveSession(userID string) (*model.Session, error)
	GetActiveSessionsByUserID(userID string) ([]*model.Session, error)
	InvalidateSession(sessionID string) error
	RefreshSession(sessionID string, newToken string, newExpiresAt time.Time) error
	TouchSession(sessionID string) error
	CleanupExpiredSessions() (int64, error)
}

type sessionService struct {
	queryBuilder mongo.QueryBuilder[model.Session]
}

func NewSessionService(db mongo.Database) SessionService {
	return &sessionService{
		queryBuilder: mongo.NewQueryBuilder[model.Session](db, model.SessionCollectionName),
	}
}

func (s *sessionService) CreateSession(
	userID string,
	token string,
	refreshToken string,
	expiresAt time.Time,
	deviceInfo model.DeviceInfo,
	userLocationInfo model.LocationInfo,
) (*model.Session, error) {

	session, err := model.NewSession(model.NewSessionArgs{
		UserId:       userID,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		DeviceInfo:   deviceInfo,
		Location:     userLocationInfo,
	})
	if err != nil {
		return nil, err
	}
	id, err := s.queryBuilder.SingleQuery().InsertOne(session)
	if err != nil {
		return nil, err
	}
	session.ID = *id
	timeNow := time.Now()
	session.CreatedAt = primitive.NewDateTimeFromTime(timeNow)
	session.UpdatedAt = primitive.NewDateTimeFromTime(timeNow)
	return session, nil
}

func (s *sessionService) GetSessionByToken(token string) (*model.Session, error) {
	filter := bson.M{"token": token, "isRevoked": false, "expiresAt": bson.M{"$gt": time.Now()}}
	options := options.FindOne().SetSort(bson.M{"expiresAt": 1})
	session, err := s.queryBuilder.SingleQuery().FilterOne(filter, options)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

func (s *sessionService) GetSessionByRefreshToken(refreshToken string) (*model.Session, error) {
	filter := bson.M{"refreshToken": refreshToken, "isRevoked": false, "expiresAt": bson.M{"$gt": time.Now()}}
	options := options.FindOne().SetSort(bson.M{"expiresAt": 1})
	session, err := s.queryBuilder.SingleQuery().FilterOne(filter, options)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

func (s *sessionService) UpdateSession(sessionID string, accessToken string, refreshToken string, expiresAt time.Time) (*model.Session, error) {
	filter := bson.M{"sessionId": sessionID, "isRevoked": false, "expiresAt": bson.M{"$gt": time.Now()}}
	update := bson.M{
		"$set": bson.M{
			"token":        accessToken,
			"refreshToken": refreshToken,
			"expiresAt":    expiresAt,
			"updatedAt":    time.Now(),
		},
	}
	session, err := s.queryBuilder.SingleQuery().FilterOneAndUpdate(filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	session.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	return session, nil
}

func (s *sessionService) UpdateSessionInfo(sessionID string, deviceInfo model.DeviceInfo, userLocation model.LocationInfo) error {
	filter := bson.M{"sessionId": sessionID, "isRevoked": false, "expiresAt": bson.M{"$gt": time.Now()}}
	update := bson.M{
		"$set": bson.M{
			"device":    deviceInfo,
			"location":  userLocation,
			"updatedAt": time.Now(),
		},
	}
	s.queryBuilder.SingleQuery().UpdateOne(filter, update, options.Update().SetUpsert(true))
	return nil
}

func (s *sessionService) GetUserActiveSession(userID string) (*model.Session, error) {
	filter := bson.M{"userId": userID, "isRevoked": false, "expiresAt": bson.M{"$gt": time.Now()}}
	options := options.FindOne().SetSort(bson.M{"expiresAt": 1})
	session, err := s.queryBuilder.SingleQuery().FilterOne(filter, options)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

func (s *sessionService) GetActiveSessionsByUserID(userID string) ([]*model.Session, error) {
	filter := bson.M{"userId": userID, "isRevoked": false, "expiresAt": bson.M{"$gt": time.Now()}}
	options := options.Find().SetSort(bson.M{"expiresAt": 1})
	sessions, err := s.queryBuilder.SingleQuery().FilterMany(filter, options)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *sessionService) InvalidateSession(sessionID string) error {
	filter := bson.M{"sessionId": sessionID}
	update := bson.M{"$set": bson.M{"isRevoked": true, "updatedAt": time.Now(), "deletedAt": time.Now()}}

	_, err := s.queryBuilder.SingleQuery().UpdateOne(filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func (s *sessionService) RefreshSession(sessionID string, newToken string, newExpiresAt time.Time) error {
	filter := bson.M{"sessionId": sessionID, "isRevoked": false, "expiresAt": bson.M{"$gt": time.Now()}}
	update := bson.M{"$set": bson.M{"token": newToken, "expiresAt": newExpiresAt, "updatedAt": time.Now()}}
	_, err := s.queryBuilder.SingleQuery().UpdateOne(filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func (s *sessionService) TouchSession(sessionID string) error {
	filter := bson.M{"sessionId": sessionID, "isRevoked": false, "expiresAt": bson.M{"$gt": time.Now()}}
	update := bson.M{"$set": bson.M{"updatedAt": time.Now()}}
	_, err := s.queryBuilder.SingleQuery().UpdateOne(filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func (s *sessionService) CleanupExpiredSessions() (int64, error) {
	filter := bson.M{"expiresAt": bson.M{"$lt": time.Now()}}
	result, err := s.queryBuilder.SingleQuery().DeleteMany(filter, options.Delete())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
