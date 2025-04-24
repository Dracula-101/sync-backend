package user

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"sync-backend/api/user/model"
	"sync-backend/arch/mongo"
	"sync-backend/utils"
)

type UserService interface {
	CreateUser(email string, password string, name string, profilePicUrl string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	ValidateUserPassword(user *model.User, password string) error
	GetUserById(id string) (*model.User, error)
	GetUserByGoogleId(googleId string) (*model.User, error)
	CreateUserWithGoogleId(googleIdToken string) (*model.User, error)
}

type userService struct {
	log              utils.AppLogger
	userQueryBuilder mongo.QueryBuilder[model.User]
}

func NewUserService(db mongo.Database) UserService {
	return &userService{
		userQueryBuilder: mongo.NewQueryBuilder[model.User](db, model.UserCollectionName),
		log:              utils.NewServiceLogger("UserService"),
	}
}

func (s *userService) CreateUser(email string, password string, name string, profilePicUrl string) (*model.User, error) {
	s.log.Debug("Creating user with email: %s", email)
	existingUser, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"email": email}, nil)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		s.log.Error("Error checking for existing user: %v", err)
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}
	if existingUser != nil {
		s.log.Error("User with this email already exists: %s", email)
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.log.Error("Error hashing password: %v", err)
		return nil, err
	}
	user, err := model.NewUser(email, hashedPassword, name, profilePicUrl)
	if err != nil {
		s.log.Error("Error creating user: %v", err)
		return nil, err
	}

	id, err := s.userQueryBuilder.SingleQuery().InsertOne(user.GetValue())
	if err != nil {
		s.log.Error("Error inserting user into database: %v", err)
		return nil, err
	}
	user.ID = *id
	return user, nil
}

func (s *userService) FindUserByEmail(email string) (*model.User, error) {
	s.log.Debug("Finding user by email: %s", email)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"email": email}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		s.log.Error("Error finding user by email: %v", err)
		return nil, err
	}
	s.log.Debug("User found: %s", user.Email)
	return user, nil
}

func (s *userService) ValidateUserPassword(user *model.User, password string) error {
	s.log.Debug("Validating password for user: %s", user.Email)
	if user.Password == nil {
		s.log.Error("User has no password set")
		return errors.New("user has no password set")
	}

	isValid, err := utils.CheckPasswordHash(password, *user.Password)
	if err != nil {
		s.log.Error("Error comparing password: %v", err)
		return fmt.Errorf("error comparing password: %v", err)
	}
	if !isValid {
		s.log.Error("Invalid password for user: %s", user.Email)
		return errors.New("invalid password")
	}
	return nil
}

func (s *userService) GetUserById(id string) (*model.User, error) {
	s.log.Debug("Getting user by ID: %s", id)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"_id": id}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, errors.New("user not found")
		}
		s.log.Error("Error getting user by ID: %v", err)
		return nil, err
	}
	s.log.Debug("User found by ID: %s", user.ID)
	return user, nil
}

func (s *userService) GetUserByGoogleId(googleId string) (*model.User, error) {
	s.log.Debug("Getting user by Google Id")
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"googleId": googleId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		s.log.Error("Error getting user by Google ID: %v", err)
		return nil, err
	}
	s.log.Debug("User found by Google ID: %s", user.Email)
	return user, nil
}

func (s *userService) CreateUserWithGoogleId(googleIdToken string) (*model.User, error) {
	s.log.Debug("Creating user with Google ID token: %s", googleIdToken[0:10]+"***********")
	googleUser, err := utils.DecodeGoogleJWTToken(googleIdToken)
	if err != nil {
		s.log.Error("Error decoding Google ID token: %v", err)
		return nil, fmt.Errorf("error decoding Google ID token: %v", err)
	}

	existingUser, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"email": googleUser.Email}, nil)

	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}

	if existingUser != nil {
		for _, provider := range existingUser.Providers {
			if provider.AuthProvider == "google" {
				s.log.Debug("User already exists with Google ID: %s", googleIdToken[0:10]+"***********")
				return existingUser, nil
			}
		}
		existingUser.Providers = append(existingUser.Providers, model.Provider{
			Id:           primitive.NewObjectID(),
			AuthIdToken:  googleIdToken,
			AuthProvider: "google",
			AddedAt:      time.Now(),
		})
		existingUser.Verified = googleUser.EmailVerified
		existingUser.ProfilePicURL = googleUser.Picture
		existingUser.Name = googleUser.Name
		existingUser.UpdatedAt = time.Now()
		_, err := s.userQueryBuilder.SingleQuery().UpdateOne(bson.M{"_id": existingUser.ID}, bson.M{
			"$set": existingUser.GetValue(),
		})
		if err != nil {
			s.log.Error("Error updating existing user: %v", err)
			return nil, fmt.Errorf("error updating existing user: %v", err)
		}
		s.log.Debug("User updated successfully: %s", existingUser.Email)
		return existingUser, nil
	} else {
		s.log.Debug("Creating new user with Google ID: %s", googleIdToken[0:10]+"***********")
		user, err := model.NewUser(googleUser.Email, googleUser.Sub, googleUser.Name, googleUser.Picture)
		if err != nil {
			return nil, err
		}
		userAuthProvider, err := model.NewAuthProvider(
			googleIdToken,
			"google",
		)
		if err != nil {
			s.log.Error("Error creating auth provider: %v", err)
			return nil, err
		}

		user.Verified = googleUser.EmailVerified
		user.Providers = append(user.Providers, *userAuthProvider)
		id, err := s.userQueryBuilder.SingleQuery().InsertOne(user.GetValue())
		if err != nil {
			s.log.Error("Error inserting user into database: %v", err)
			return nil, err
		}
		user.ID = *id
		s.log.Debug("User created successfully: %s", user.Email)
		return user, nil
	}
}
