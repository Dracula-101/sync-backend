package user

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"sync-backend/api/common/media"
	"sync-backend/api/user/model"
	"sync-backend/arch/common"
	"sync-backend/arch/mongo"
	"sync-backend/utils"
)

type UserService interface {
	/* CREATING USER */
	CreateUser(userName string, email string, password string, profile string, backgroundPic string, locale string, timezone string, country string) (*model.User, error)
	CreateUserWithGoogleId(userName string, googleIdToken string, locale string, timezone string, country string) (*model.User, error)

	/* FINDING USER */
	FindUserById(userId string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	FindUserAuthProvider(userId string, username string, providerName string) (*model.User, error)

	/* USER INFO UPDATE */
	UpdateLoginHistory(userId string, loginHistory model.LoginHistory) error

	/* USER AUTHENTICATION */
	ValidateUserPassword(user *model.User, password string) error

	/* USER COMMUNITY */
	JoinCommunity(userId string, communityId string) error
	LeaveCommunity(userId string, communityId string) error

	/* USER FOLLOWING */
	FollowUser(userId string, followUserId string) error
	UnfollowUser(userId string, unfollowUserId string) error
	BlockUser(userId string, blockUserId string) error
	UnblockUser(userId string, unblockUserId string) error

	/* MODERATOR */
	AddModerator(userId string, communityId string) error
	RemoveModerator(userId string, communityId string) error
}

type userService struct {
	mediaService       media.MediaService
	log                utils.AppLogger
	userQueryBuilder   mongo.QueryBuilder[model.User]
	transactionBuilder mongo.TransactionBuilder
}

func NewUserService(db mongo.Database, mediaService media.MediaService) UserService {
	return &userService{
		mediaService:       mediaService,
		userQueryBuilder:   mongo.NewQueryBuilder[model.User](db, model.UserCollectionName),
		transactionBuilder: mongo.NewTransactionBuilder(db),
		log:                utils.NewServiceLogger("UserService"),
	}
}

func (s *userService) CreateUser(userName string, email string, password string, profile string, backgroundPic string, locale string, timezone string, country string) (*model.User, error) {
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
		return nil, NewDBError("checking for existing user", err.Error())
	}

	if existingUser != nil {
		if existingUser.Email == email {
			s.log.Error("User with this email already exists: %s", email)
			return nil, NewUserExistsByEmailError(email)
		} else {
			s.log.Error("User with this username already exists: %s", userName)
			return nil, NewUserExistsByUsernameError(userName)
		}
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.log.Error("Error hashing password: %v", err)
		return nil, NewDBError("hashing password", err.Error())
	}
	var profilePicUrl, profileId string
	var profileHeight, profileWidth int
	var backgroundUrl, backgroundId string
	var backgroundHeight, backgroundWidth int
	if profile != "" {
		profileInfo, _ := s.mediaService.UploadMedia(profile, userName+"_profile", "profile")
		profileId = profileInfo.Id
		profilePicUrl = profileInfo.Url
		profileHeight = profileInfo.Height
		profileWidth = profileInfo.Width
	} else {
		profile = "https://placehold.co/150x150.png"
		profileId = "default-profile-id"
		profilePicUrl = profile
		profileHeight = 150
		profileWidth = 150
	}

	if backgroundPic != "" {
		backgroundInfo, _ := s.mediaService.UploadMedia(backgroundPic, userName+"_background", "background")
		backgroundId = backgroundInfo.Id
		backgroundUrl = backgroundInfo.Url
		backgroundHeight = backgroundInfo.Height
		backgroundWidth = backgroundInfo.Width
	} else {
		backgroundPic = "https://placehold.co/1200x400.png"
		backgroundId = "default-background-id"
		backgroundUrl = backgroundPic
		backgroundHeight = 400
		backgroundWidth = 1200
	}

	user, err := model.NewUser(model.NewUserArgs{
		UserName:     userName,
		Email:        email,
		PasswordHash: hashedPassword,
		AvatarUrl: model.Image{
			Id:     profileId,
			Url:    profilePicUrl,
			Width:  profileWidth,
			Height: profileHeight,
		},
		BackgroundUrl: model.Image{
			Id:     backgroundId,
			Url:    backgroundUrl,
			Width:  backgroundWidth,
			Height: backgroundHeight,
		},
		Language:    common.GetLanguageByID(locale),
		TimeZone:    common.GetTimeZone(timezone),
		Theme:       "light",
		Country:     country,
		DeviceToken: *model.NewDeviceToken("default-token-id-here", "DEVICE_ID", "PUSH"),
	})
	if err != nil {
		s.log.Error("Error creating user: %v", err)
		return nil, NewDBError("creating user", err.Error())
	}

	id, err := s.userQueryBuilder.SingleQuery().InsertOne(user.GetValue())
	if err != nil {
		s.log.Error("Error inserting user into database: %v", err)
		return nil, NewDBError("inserting user into database", err.Error())
	}
	user.Id = *id
	return user, nil
}

func (s *userService) CreateUserWithGoogleId(userName string, googleIdToken string, locale string, timezone string, country string) (*model.User, error) {
	s.log.Debug("Creating user with Google ID token: %s", googleIdToken[0:10]+"***********")
	googleUser, err := utils.DecodeGoogleJWTToken(googleIdToken)
	if err != nil {
		s.log.Error("Error decoding Google ID token: %v", err)
		return nil, NewDBError("decoding Google ID token", err.Error())
	}

	existingUser, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"email": googleUser.Email}, nil)

	if err != nil && !mongo.IsNoDocumentFoundError(err) {
		return nil, NewDBError("checking for existing user", err.Error())
	}
	width, height, _ := utils.GetImageSize(googleUser.Picture)
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
		existingUser.Avatar.Profile.Url = googleUser.Picture
		existingUser.Avatar.Profile.Width = width
		existingUser.Avatar.Profile.Height = height
		existingUser.Email = googleUser.Email
		existingUser.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
		_, err := s.userQueryBuilder.SingleQuery().UpdateOne(bson.M{"userId": existingUser.UserId}, bson.M{
			"$set": existingUser.GetValue(),
		}, nil)
		if err != nil {
			s.log.Error("Error updating existing user: %v", err)
			return nil, NewDBError("updating existing user", err.Error())
		}
		s.log.Debug("User updated successfully: %s", existingUser.Email)
		return existingUser, nil
	} else {
		s.log.Debug("Creating new user with Google ID: %s", googleIdToken[0:10]+"***********")
		user, err := model.NewUser(model.NewUserArgs{
			UserName: userName,
			Email:    googleUser.Email,
			AvatarUrl: model.Image{
				Id:     "default-profile-id",
				Url:    googleUser.Picture,
				Width:  width,
				Height: height,
			},
			BackgroundUrl: model.Image{
				Id:     "default-background-id",
				Url:    "https://placehold.co/1200x400.png",
				Width:  1200,
				Height: 400,
			},
			Language:    common.GetLanguageByID(locale),
			TimeZone:    common.GetTimeZone(timezone),
			DeviceToken: *model.NewDeviceToken("default-token-id-here", "DEVICE_ID", "PUSH"),
		})
		if err != nil {
			return nil, NewDBError("creating user from Google ID", err.Error())
		}
		userAuthProvider, err := model.NewAuthProvider(
			googleIdToken,
			model.GoogleProviderName,
			fmt.Sprintf("%s %s", googleUser.GivenName, googleUser.FamilyName),
		)
		if err != nil {
			s.log.Error("Error creating auth provider: %v", err)
			return nil, NewDBError("creating auth provider", err.Error())
		}

		user.VerifiedEmail = googleUser.EmailVerified
		user.Providers = append(user.Providers, *userAuthProvider)
		id, err := s.userQueryBuilder.SingleQuery().InsertOne(user.GetValue())
		if err != nil {
			s.log.Error("Error inserting user into database: %v", err)
			return nil, NewDBError("inserting user into database", err.Error())
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
			return nil, NewUserNotFoundError(userId)
		}
		s.log.Error("Error getting user by ID: %v", err)
		return nil, NewDBError("getting user by ID", err.Error())
	}
	if user.Status == model.Deleted {
		s.log.Error("User is deleted: %s", userId)
		return nil, NewUserDeletedError(userId)
	}
	if user.Status == model.Inactive {
		s.log.Error("User is inactive: %s", userId)
		return nil, NewUserInactiveError(userId)
	}
	if user.Status == model.Banned {
		s.log.Error("User is banned: %s", userId)
		return nil, NewUserBannedError(userId)
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
		return nil, NewDBError("finding user by email", err.Error())
	}
	if user.Status == model.Deleted {
		s.log.Error("User is deleted: %s", email)
		return nil, NewUserDeletedError(email)
	}
	if user.Status == model.Inactive {
		s.log.Error("User is inactive: %s", email)
		return nil, NewUserInactiveError(email)
	}
	if user.Status == model.Banned {
		s.log.Error("User is banned: %s", email)
		return nil, NewUserBannedError(email)
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
		return nil, NewDBError("finding user by username", err.Error())
	}
	if user.Status == model.Deleted {
		s.log.Error("User is deleted: %s", username)
		return nil, NewUserDeletedError(username)
	}
	if user.Status == model.Inactive {
		s.log.Error("User is inactive: %s", username)
		return nil, NewUserInactiveError(username)
	}
	if user.Status == model.Banned {
		s.log.Error("User is banned: %s", username)
		return nil, NewUserBannedError(username)
	}
	s.log.Debug("User found: %s", user.Username)
	return user, nil
}

func (s *userService) FindUserAuthProvider(userId string, username string, providerName string) (*model.User, error) {
	s.log.Debug("Finding auth provider by user ID: %s and provider name: %s", userId, providerName)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId, "username": username, "providers.providerName": providerName}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		s.log.Error("Error finding auth provider by user ID: %v", err)
		return nil, NewDBError("finding auth provider by user ID", err.Error())
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

func (s *userService) UpdateLoginHistory(userId string, loginHistory model.LoginHistory) error {
	s.log.Debug("Updating login history for user ID: %s", userId)

	result, err := s.userQueryBuilder.SingleQuery().UpdateOne(bson.M{"userId": userId}, bson.M{
		"$push": bson.M{
			"loginHistory": bson.M{
				"$each":     []model.LoginHistory{loginHistory},
				"$slice":    -10,
				"$position": 0,
			},
		},
		"$set": bson.M{
			"lastLogin": loginHistory.LoginTime,
		},
	}, nil)
	if err != nil {
		s.log.Error("Error updating login history: %v", err)
		return NewDBError("updating login history", err.Error())
	}

	s.log.Debug("Login history updated successfully for user ID: %s - Modified count: %d", userId, result.ModifiedCount)
	return nil
}

func (s *userService) ValidateUserPassword(user *model.User, password string) error {
	s.log.Debug("Validating password for user: %s", user.Email)

	isValid, err := utils.CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		s.log.Error("Error comparing password: %v", err)
		return NewDBError("comparing password", err.Error())
	}
	if !isValid {
		s.log.Error("Invalid password for user: %s", user.Email)
		return NewForbiddenUserActionError("validate password for", user.Email, user.Email)
	}
	return nil
}

func (s *userService) JoinCommunity(userId string, communityId string) error {
	s.log.Debug("Joining community %s for user %s", communityId, userId)
	_, err := s.userQueryBuilder.SingleQuery().UpdateOne(bson.M{"userId": userId}, bson.M{
		"$addToSet": bson.M{
			"joinedWavelengths": communityId,
		},
	}, nil)
	if err != nil {
		s.log.Error("Error joining community: %v", err)
		return NewDBError("joining community", err.Error())
	}

	s.log.Debug("User %s joined community %s successfully", userId, communityId)
	return nil
}

func (s *userService) LeaveCommunity(userId string, communityId string) error {
	s.log.Debug("Leaving community %s for user %s", communityId, userId)
	_, err := s.userQueryBuilder.SingleQuery().UpdateOne(bson.M{"userId": userId}, bson.M{
		"$pull": bson.M{
			"joinedWavelengths": communityId,
		},
	}, nil)
	if err != nil {
		s.log.Error("Error leaving community: %v", err)
		return NewDBError("leaving community", err.Error())
	}

	s.log.Debug("User %s left community %s successfully", userId, communityId)
	return nil
}

func (s *userService) FollowUser(userId string, followUserId string) error {
	s.log.Debug("Following user %s for user %s", followUserId, userId)
	_, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": followUserId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User to follow does not exist: %s", followUserId)
			return NewUserNotFoundError(followUserId)
		}
		s.log.Error("Error checking if user to follow exists: %v", err)
		return NewDBError("checking if user to follow exists", err.Error())
	}
	if userId == followUserId {
		s.log.Error("Cannot follow self")
		return NewSelfActionError("follow")
	}

	transaction := s.transactionBuilder.GetTransaction(time.Minute * 5)
	if err := transaction.Start(); err != nil {
		s.log.Error("Error starting transaction: %v", err)
		return NewDBError("starting transaction", err.Error())
	}
	err = transaction.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		userCollection := session.Collection(model.UserCollectionName)
		_, err = userCollection.UpdateOne(
			bson.M{"userId": userId},
			bson.M{
				"$addToSet": bson.M{
					"follows": followUserId,
				},
			},
		)
		if err != nil {
			s.log.Error("error following user: %v", err)
			return NewDBError("following user", err.Error())
		}
		_, err = userCollection.UpdateOne(
			bson.M{"userId": followUserId},
			bson.M{
				"$addToSet": bson.M{
					"followers": userId,
				},
			},
		)
		if err != nil {
			s.log.Error("error following user: %v", err)
			return NewDBError("following user", err.Error())
		}
		return nil
	})

	if err != nil {
		s.log.Error("Error following user: %v", err)
		return NewDBError("following user", err.Error())
	}

	s.log.Debug("User %s followed user %s successfully", userId, followUserId)
	return nil
}

func (s *userService) UnfollowUser(userId string, unfollowUserId string) error {
	s.log.Debug("Unfollowing user %s for user %s", unfollowUserId, userId)
	_, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": unfollowUserId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User to follow does not exist: %s", unfollowUserId)
			return NewUserNotFoundError(unfollowUserId)
		}
		s.log.Error("Error checking if user to follow exists: %v", err)
		return NewDBError("checking if user to follow exists", err.Error())
	}
	if userId == unfollowUserId {
		s.log.Error("Cannot follow self")
		return NewSelfActionError("unfollow")
	}

	transaction := s.transactionBuilder.GetTransaction(time.Minute * 5)
	if err := transaction.Start(); err != nil {
		s.log.Error("Error starting transaction: %v", err)
		return NewDBError("starting transaction", err.Error())
	}

	err = transaction.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		userCollection := session.Collection(model.UserCollectionName)
		_, err = userCollection.UpdateOne(
			bson.M{"userId": userId},
			bson.M{
				"$pull": bson.M{
					"follows": unfollowUserId,
				},
			},
		)
		if err != nil {
			s.log.Error("error unfollowing user: %v", err)
			return NewDBError("unfollowing user", err.Error())
		}
		_, err = userCollection.UpdateOne(
			bson.M{"userId": unfollowUserId},
			bson.M{
				"$pull": bson.M{
					"followers": userId,
				},
			},
		)
		if err != nil {
			s.log.Error("error unfollowing user: %v", err)
			return NewDBError("unfollowing user", err.Error())
		}
		return nil
	})

	if err != nil {
		s.log.Error("Error unfollowing user: %v", err)
		return NewDBError("unfollowing user", err.Error())
	}

	s.log.Debug("User %s unfollowed user %s successfully", userId, unfollowUserId)
	return nil
}

func (s *userService) BlockUser(userId string, blockUserId string) error {
	s.log.Debug("Blocking user %s for user %s", blockUserId, userId)
	_, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": blockUserId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User to block does not exist: %s", blockUserId)
			return NewUserNotFoundError(blockUserId)
		}
		s.log.Error("Error checking if user to block exists: %v", err)
		return NewDBError("checking if user to block exists", err.Error())
	}
	if userId == blockUserId {
		s.log.Error("Cannot block self")
		return NewSelfActionError("block")
	}
	transaction := s.transactionBuilder.GetTransaction(time.Minute * 5)
	if err := transaction.Start(); err != nil {
		s.log.Error("Error starting transaction: %v", err)
		return NewDBError("starting transaction", err.Error())
	}

	err = transaction.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		userCollection := session.Collection(model.UserCollectionName)
		_, err = userCollection.UpdateOne(
			bson.M{"userId": userId},
			bson.M{"$addToSet": bson.M{
				"preferences.blockList": blockUserId,
			},
			},
		)
		if err != nil {
			s.log.Error("error blocking user: %v", err)
			return NewDBError("blocking user", err.Error())
		}
		return nil
	})

	if err != nil {
		return NewDBError("blocking user", err.Error())
	}

	s.log.Debug("User %s blocked user %s successfully", userId, blockUserId)
	return nil
}

func (s *userService) UnblockUser(userId string, unblockUserId string) error {
	s.log.Debug("Unblocking user %s for user %s", unblockUserId, userId)
	_, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": unblockUserId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User to unblock does not exist: %s", unblockUserId)
			return NewUserNotFoundError(unblockUserId)
		}
		s.log.Error("Error checking if user to unblock exists: %v", err)
		return NewDBError("checking if user to unblock exists", err.Error())
	}
	if userId == unblockUserId {
		s.log.Error("Cannot block self")
		return NewSelfActionError("unblock")
	}
	transaction := s.transactionBuilder.GetTransaction(time.Minute * 5)
	if err := transaction.Start(); err != nil {
		s.log.Error("Error starting transaction: %v", err)
		return NewDBError("starting transaction", err.Error())
	}

	err = transaction.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		userCollection := session.Collection(model.UserCollectionName)
		_, err = userCollection.UpdateOne(
			bson.M{"userId": userId},
			bson.M{
				"$pull": bson.M{
					"preferences.blockList": unblockUserId,
				},
			},
		)
		if err != nil {
			s.log.Error("error unblocking user: %v", err)
			return NewDBError("unblocking user", err.Error())
		}
		return nil
	})

	if err != nil {
		s.log.Error("Error unblocking user: %v", err)
		return NewDBError("unblocking user", err.Error())
	}

	s.log.Debug("User %s unblocked user %s successfully", userId, unblockUserId)
	return nil
}

func (s *userService) AddModerator(userId string, communityId string) error {
	s.log.Debug("Adding moderator %s to community %s", userId, communityId)
	_, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User to add as moderator does not exist: %s", userId)
			return NewUserNotFoundError(userId)
		}
		s.log.Error("Error checking if user exists: %v", err)
		return NewDBError("checking if user exists", err.Error())
	}

	transaction := s.transactionBuilder.GetTransaction(time.Minute * 5)
	if err := transaction.Start(); err != nil {
		s.log.Error("Error starting transaction: %v", err)
		return NewDBError("starting transaction", err.Error())
	}

	err = transaction.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		userCollection := session.Collection(model.UserCollectionName)
		_, err = userCollection.UpdateOne(
			bson.M{"userId": userId},
			bson.M{
				"$addToSet": bson.M{
					"moderatedCommunities": communityId,
				},
			},
		)
		if err != nil {
			s.log.Error("error adding moderator: %v", err)
			return NewDBError("adding moderator", err.Error())
		}
		return nil
	})

	if err != nil {
		s.log.Error("Error adding moderator: %v", err)
		return NewDBError("adding moderator", err.Error())
	}

	s.log.Debug("Moderator %s added to community %s successfully", userId, communityId)
	return nil
}

func (s *userService) RemoveModerator(userId string, communityId string) error {
	s.log.Debug("Removing moderator %s from community %s", userId, communityId)
	_, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User to remove as moderator does not exist: %s", userId)
			return NewUserNotFoundError(userId)
		}
		s.log.Error("Error checking if user exists: %v", err)
		return NewDBError("checking if user exists", err.Error())
	}
	transaction := s.transactionBuilder.GetTransaction(time.Minute * 5)
	if err := transaction.Start(); err != nil {
		s.log.Error("Error starting transaction: %v", err)
		return NewDBError("starting transaction", err.Error())
	}
	err = transaction.PerformSingleTransaction(func(session mongo.TransactionSession) error {
		userCollection := session.Collection(model.UserCollectionName)
		_, err = userCollection.UpdateOne(
			bson.M{"userId": userId},
			bson.M{
				"$pull": bson.M{
					"moderatedCommunities": communityId,
				},
			},
		)
		if err != nil {
			s.log.Error("error removing moderator: %v", err)
			return NewDBError("removing moderator", err.Error())
		}
		return nil
	})
	if err != nil {
		s.log.Error("Error removing moderator: %v", err)
		return NewDBError("removing moderator", err.Error())
	}
	s.log.Debug("Moderator %s removed from community %s successfully", userId, communityId)
	return nil
}
