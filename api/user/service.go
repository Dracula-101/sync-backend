package user

import (
	"fmt"
	"sync-backend/api/user/model"
	"sync-backend/arch/mongo"
	"sync-backend/utils"

	"go.mongodb.org/mongo-driver/bson"
)

type UserService interface {
	CreateUser(email string, password string, name string, profilePicUrl string) (*model.User, error)
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
	existingUser, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"email": email})
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("a user with the email %s already exists", email)
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
		fmt.Printf("Error inserting user: %v\n", err)
		return nil, err
	}
	user.ID = *id
	return user, nil
}
