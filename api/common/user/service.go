package user

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"sync-backend/api/common/user/model"
	"sync-backend/arch/common"
	"sync-backend/arch/mongo"
	"sync-backend/utils"
)

type UserService interface {
	/* CREATING USER */
	CreateUser(userName string, email string, password string, profilePicUrl string) (*model.User, error)
	CreateUserWithGoogleId(userName string, googleIdToken string) (*model.User, error)

	/* FINDING USER */
	FindUserById(userId string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	FindUserAuthProvider(userId string, providerName string) (*model.User, error)

	/* USER AUTHENTICATION */
	ValidateUserPassword(user *model.User, password string) error
}

type userService struct {
	log                utils.AppLogger
	userQueryBuilder   mongo.QueryBuilder[model.User]
	transactionBuilder mongo.TransactionBuilder
}

func NewUserService(db mongo.Database) UserService {
	return &userService{
		userQueryBuilder:   mongo.NewQueryBuilder[model.User](db, model.UserCollectionName),
		transactionBuilder: mongo.NewTransactionBuilder(db),
		log:                utils.NewServiceLogger("UserService"),
	}
}

func (s *userService) CreateUser(userName string, email string, password string, profilePicUrl string) (*model.User, error) {
	s.log.Debug("Creating user with email: %s", email)
	filter := bson.M{
		"$or": []bson.M{
			{"email": email},
			{"username": userName},
		},
	}

	existingUser, err := s.userQueryBuilder.SingleQuery().FilterOne(filter, nil)
	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		s.log.Error("Error checking for existing user: %v", err)
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}

	if existingUser != nil {
		if existingUser.Email == email {
			s.log.Error("User with this email already exists: %s", email)
			return nil, errors.New("user with this email already exists")
		} else {
			s.log.Error("User with this username already exists: %s", userName)
			return nil, errors.New("user with this username already exists")
		}
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.log.Error("Error hashing password: %v", err)
		return nil, err
	}
	user, err := model.NewUser(model.NewUserArgs{
		UserName:      userName,
		Email:         email,
		PasswordHash:  hashedPassword,
		AvatarUrl:     profilePicUrl,
		BackgroundUrl: "https://placehold.co/1200x400.png",
		Language:      common.English,
		TimeZone:      common.AsiaKolkata,
		DeviceToken:   *model.NewDeviceToken("default-token-id-here", "DEVICE_ID", "PUSH"),
	})
	if err != nil {
		s.log.Error("Error creating user: %v", err)
		return nil, err
	}

	id, err := s.userQueryBuilder.SingleQuery().InsertOne(user.GetValue())
	if err != nil {
		s.log.Error("Error inserting user into database: %v", err)
		return nil, err
	}
	user.Id = *id
	return user, nil
}

func (s *userService) CreateUserWithGoogleId(userName string, googleIdToken string) (*model.User, error) {
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
			if provider.AuthProvider == model.GoogleProviderName {
				s.log.Debug("User already exists with Google ID: %s", googleIdToken[0:10]+"***********")
				return existingUser, nil
			}
		}
		existingUser.Providers = append(existingUser.Providers, model.Provider{
			Id:           primitive.NewObjectID(),
			AuthIdToken:  googleIdToken,
			AuthProvider: model.GoogleProviderName,
			AddedAt:      time.Now(),
		})
		existingUser.VerifiedEmail = googleUser.EmailVerified
		existingUser.Avatar.ProfilePic.Url = googleUser.Picture
		existingUser.Email = googleUser.Email
		existingUser.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
		_, err := s.userQueryBuilder.SingleQuery().UpdateOne(bson.M{"userId": existingUser.UserId}, bson.M{
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
		user, err := model.NewUser(model.NewUserArgs{
			UserName:      userName,
			Email:         googleUser.Email,
			PasswordHash:  googleIdToken,
			AvatarUrl:     googleUser.Picture,
			BackgroundUrl: "https://placehold.co/1200x400.png",
			Language:      common.English,
			TimeZone:      common.AsiaKolkata,
			DeviceToken:   *model.NewDeviceToken("default-token-id-here", "DEVICE_ID", "PUSH"),
		})
		if err != nil {
			return nil, err
		}
		userAuthProvider, err := model.NewAuthProvider(
			googleIdToken,
			model.GoogleProviderName,
			fmt.Sprintf("%s %s", googleUser.GivenName, googleUser.FamilyName),
		)
		if err != nil {
			s.log.Error("Error creating auth provider: %v", err)
			return nil, err
		}

		user.VerifiedEmail = googleUser.EmailVerified
		user.Providers = append(user.Providers, *userAuthProvider)
		id, err := s.userQueryBuilder.SingleQuery().InsertOne(user.GetValue())
		if err != nil {
			s.log.Error("Error inserting user into database: %v", err)
			return nil, err
		}
		s.log.Debug("User created successfully: %s - %s", user.Email, id.Hex())
		return user, nil
	}
}

func (s *userService) FindUserById(userId string) (*model.User, error) {
	s.log.Debug("Getting user by ID: %s", userId)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, errors.New("user not found")
		}
		s.log.Error("Error getting user by ID: %v", err)
		return nil, err
	}
	s.log.Debug("User found by ID: %s", user.UserId)
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

func (s *userService) FindUserByUsername(username string) (*model.User, error) {
	s.log.Debug("Finding user by username: %s", username)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"username": username}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		s.log.Error("Error finding user by username: %v", err)
		return nil, err
	}
	s.log.Debug("User found: %s", user.Username)
	return user, nil
}

func (s *userService) FindUserAuthProvider(userId string, providerName string) (*model.User, error) {
	s.log.Debug("Finding auth provider by user ID: %s and provider name: %s", userId, providerName)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId, "providers.providerName": providerName}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		s.log.Error("Error finding auth provider by user ID: %v", err)
		return nil, err
	}
	if user == nil {
		s.log.Debug("No auth provider found for user ID: %s and provider name: %s", userId, providerName)
		return nil, nil
	}

	for _, p := range user.Providers {
		if p.AuthProvider == providerName {
			s.log.Debug("Auth provider found: %s", p.AuthProvider)
			return user, nil
		}
	}
	s.log.Debug("No auth provider found for user ID: %s and provider name: %s", userId, providerName)
	return nil, nil
}

func (s *userService) ValidateUserPassword(user *model.User, password string) error {
	s.log.Debug("Validating password for user: %s", user.Email)

	isValid, err := utils.CheckPasswordHash(password, user.PasswordHash)
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
