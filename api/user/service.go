package user

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"sync-backend/api/user/model"
	"sync-backend/arch/mongo"
	"sync-backend/utils"
)

type UserService interface {
	CreateUser(email string, password string, name string, profilePicUrl string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	ValidateUserPassword(user *model.User, password string) error
	GetUserById(id string) (*model.User, error)
}

type userService struct {
	userQueryBuilder mongo.QueryBuilder[model.User]
}

func NewUserService(db mongo.Database) UserService {
	return &userService{
		userQueryBuilder: mongo.NewQueryBuilder[model.User](db, model.UserCollectionName),
	}
}

func (s *userService) CreateUser(email string, password string, name string, profilePicUrl string) (*model.User, error) {
	// Check if a user with the same email already exists
	existingUser, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"email": email}, nil)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user, err := model.NewUser(email, hashedPassword, name, profilePicUrl)
	if err != nil {
		return nil, err
	}

	id, err := s.userQueryBuilder.SingleQuery().InsertOne(user.GetValue())
	if err != nil {
		return nil, err
	}
	user.ID = *id
	return user, nil
}

func (s *userService) FindUserByEmail(email string) (*model.User, error) {
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"email": email}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) ValidateUserPassword(user *model.User, password string) error {
	if user.Password == nil {
		return errors.New("user has no password set")
	}

	isValid, err := utils.CheckPasswordHash(password, *user.Password)
	if err != nil {
		return fmt.Errorf("error comparing password: %v", err)
	}
	if !isValid {
		return errors.New("invalid password")
	}
	return nil
}

func (s *userService) GetUserById(id string) (*model.User, error) {
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"_id": id}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
