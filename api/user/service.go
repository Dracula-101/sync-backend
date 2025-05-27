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
	"sync-backend/arch/network"
	"sync-backend/utils"
)

const EMPTY_PASSWORD_HASH = "$2a$10$Cv/Xb2ykZ9FLmWyB6vaPEueAzA51kkU2GDZj8C4hwgAH3gQhwIo.q"

type UserService interface {
	/* CREATING USER */
	CreateUser(userName string, email string, password string, profile string, backgroundPic string, locale string, timezone string, country string) (*model.User, network.ApiError)
	CreateUserWithGoogleId(userName string, googleIdToken string, locale string, timezone string, country string) (*model.User, network.ApiError)

	/* FINDING USER */
	FindUserById(userId string) (*model.User, network.ApiError)
	FindUserByEmail(email string) (*model.User, network.ApiError)
	FindUserByUsername(username string) (*model.User, network.ApiError)
	FindUserAuthProvider(userId string, username string, providerName string) (*model.User, network.ApiError)

	/* USER INFO UPDATE */
	UpdateUserProfile(userId string, bio *string, profilePicPath *string, backgroundPicPath *string) (*model.User, network.ApiError)
	UpdateUserPreferences(userId string, preferences model.UserPreferences) (*model.User, network.ApiError)
	UpdateLoginHistory(userId string, loginHistory model.LoginHistory) network.ApiError

	/* USER FUNCTIONALITY */
	ValidateUserPassword(user *model.User, password string) network.ApiError
	DeleteUser(userId string) network.ApiError
	ChangePassword(userId string, oldPassword string, newPassword string) network.ApiError

	/* USER COMMUNITY */
	JoinCommunity(userId string, communityId string) network.ApiError
	LeaveCommunity(userId string, communityId string) network.ApiError
	SearchUsers(userId string, query string, page int, limit int) ([]*model.SearchUser, network.ApiError)

	/* USER FOLLOWING */
	FollowUser(userId string, followUserId string) network.ApiError
	UnfollowUser(userId string, unfollowUserId string) network.ApiError
	BlockUser(userId string, blockUserId string) network.ApiError
	UnblockUser(userId string, unblockUserId string) network.ApiError

	/* MODERATOR */
	AddModerator(userId string, communityId string) network.ApiError
	RemoveModerator(userId string, communityId string) network.ApiError
}

type userService struct {
	mediaService          media.MediaService
	log                   utils.AppLogger
	userQueryBuilder      mongo.QueryBuilder[model.User]
	transactionBuilder    mongo.TransactionBuilder
	searchUsersAggregator mongo.AggregateBuilder[model.User, model.SearchUser]
}

func NewUserService(db mongo.Database, mediaService media.MediaService) UserService {
	return &userService{
		mediaService:          mediaService,
		log:                   utils.NewServiceLogger("UserService"),
		userQueryBuilder:      mongo.NewQueryBuilder[model.User](db, model.UserCollectionName),
		transactionBuilder:    mongo.NewTransactionBuilder(db),
		searchUsersAggregator: mongo.NewAggregateBuilder[model.User, model.SearchUser](db, model.UserCollectionName),
	}
}

func (s *userService) CreateUser(userName string, email string, password string, profile string, backgroundPic string, locale string, timezone string, country string) (*model.User, network.ApiError) {
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

func (s *userService) CreateUserWithGoogleId(userName string, googleIdToken string, locale string, timezone string, country string) (*model.User, network.ApiError) {
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

func (s *userService) FindUserById(userId string) (*model.User, network.ApiError) {
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
		return nil, NewUserBannedError(userId, "User violated terms and conditions of the platform")
	}

	s.log.Debug("User found by ID: %s", user.UserId)
	return user, nil
}

func (s *userService) FindUserByEmail(email string) (*model.User, network.ApiError) {
	s.log.Debug("Finding user by email: %s", email)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"email": email}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			return nil, nil
		}
		s.log.Error("Error finding user by email: %v", err)
		return nil, NewDBError("finding user by email", err.Error())
	}

	switch user.Status {
	case model.Deleted:
		s.log.Error("User is deleted: %s", email)
		return nil, NewUserDeletedError(email)
	case model.Inactive:
		s.log.Error("User is inactive: %s", email)
		return nil, NewUserInactiveError(email)
	case model.Banned:
		s.log.Error("User is banned: %s", email)
		return nil, NewUserBannedError(email, "User violated terms and conditions of the platform")
	}

	if user == nil {
		s.log.Debug("No user found with email: %s", email)
		return nil, NewUserNotFoundByEmailError(email)
	}

	s.log.Debug("User found: %s", user.Email)
	return user, nil
}

func (s *userService) FindUserByUsername(username string) (*model.User, network.ApiError) {
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
		return nil, NewUserBannedError(username, "User violated terms and conditions of the platform")
	}
	s.log.Debug("User found: %s", user.Username)
	return user, nil
}

func (s *userService) FindUserAuthProvider(userId string, username string, providerName string) (*model.User, network.ApiError) {
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

func (s *userService) UpdateUserPreferences(userId string, preferences model.UserPreferences) (*model.User, network.ApiError) {
	s.log.Debug("Updating user preferences for user ID: %s", userId)

	// Assuming the user is already fetched and available as 'user'
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User not found for profile update: %s", userId)
			return nil, NewUserNotFoundError(userId)
		}
		s.log.Error("Error fetching user for profile update: %v", err)
		return nil, NewDBError("fetching user for profile update", err.Error())
	}

	switch user.Status {
	case model.Deleted:
		s.log.Error("User is deleted: %s", userId)
		return nil, NewUserDeletedError(userId)
	case model.Inactive:
		s.log.Error("User is inactive: %s", userId)
		return nil, NewUserInactiveError(userId)
	case model.Banned:
		s.log.Error("User is banned: %s", userId)
		return nil, NewUserBannedError(userId, "User violated terms and conditions of the platform")
	}

	user.Preferences = preferences
	user.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	// update only if not nil
	_, err = s.userQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"userId": user.UserId},
		bson.M{"$set": bson.M{
			"preferences": user.Preferences,
			"updatedAt":   user.UpdatedAt,
		}},
		nil,
	)
	if err != nil {
		s.log.Error("Error updating user preferences: %v", err)
		return nil, NewDBError("updating user preferences", err.Error())
	}
	s.log.Debug("User preferences updated successfully for user ID: %s", user.UserId)
	return user, nil
}

func (s *userService) UpdateUserProfile(userId string, bio *string, profilePicPath *string, backgroundPicPath *string) (*model.User, network.ApiError) {
	s.log.Debug("Updating user profile for user ID: %s", userId)

	// Assuming the user is already fetched and available as 'user'
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User not found for profile update: %s", userId)
			return nil, NewUserNotFoundError(userId)
		}
		s.log.Error("Error fetching user for profile update: %v", err)
		return nil, NewDBError("fetching user for profile update", err.Error())
	}

	switch user.Status {
	case model.Deleted:
		s.log.Error("User is deleted: %s", userId)
		return nil, NewUserDeletedError(userId)
	case model.Inactive:
		s.log.Error("User is inactive: %s", userId)
		return nil, NewUserInactiveError(userId)
	case model.Banned:
		s.log.Error("User is banned: %s", userId)
		return nil, NewUserBannedError(userId, "User violated terms and conditions of the platform")
	}

	if profilePicPath != nil {
		s.mediaService.DeleteMedia(user.Avatar.Profile.Id) // Delete old profile pic if exists
		profileInfo, _ := s.mediaService.UploadMedia(*profilePicPath, user.Username+"_profile", "profile")
		user.Avatar.Profile = model.Image{
			Id:     profileInfo.Id,
			Url:    profileInfo.Url,
			Width:  profileInfo.Width,
			Height: profileInfo.Height,
		}
	}

	if backgroundPicPath != nil {
		s.mediaService.DeleteMedia(user.Avatar.Background.Id) // Delete old background if exists
		backgroundInfo, _ := s.mediaService.UploadMedia(*backgroundPicPath, user.Username+"_background", "background")
		user.Avatar.Background = model.Image{
			Id:     backgroundInfo.Id,
			Url:    backgroundInfo.Url,
			Width:  backgroundInfo.Width,
			Height: backgroundInfo.Height,
		}
	}

	if bio != nil {
		user.Bio = *bio
	}
	user.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	_, err = s.userQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"userId": user.UserId},
		bson.M{"$set": user.GetValue()},
		nil,
	)
	if err != nil {
		s.log.Error("Error updating user profile: %v", err)
		return nil, NewDBError("updating user profile", err.Error())
	}

	s.log.Debug("User profile updated successfully for user ID: %s", user.UserId)
	return user, nil
}

func (s *userService) UpdateLoginHistory(userId string, loginHistory model.LoginHistory) network.ApiError {
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

func (s *userService) ValidateUserPassword(user *model.User, password string) network.ApiError {
	s.log.Debug("Validating password for user: %s", user.Email)

	isValid, err := utils.CheckPasswordHash(user.PasswordHash, password)
	if err != nil {
		s.log.Error("Error comparing password: %v", err)
		return NewDBError("comparing password", err.Error())
	}
	if !isValid {
		s.log.Error("Invalid password for user: %s", user.Email)
		return NewWrongPasswordError(user.UserId)
	}
	return nil
}

func (s *userService) DeleteUser(userId string) network.ApiError {
	s.log.Debug("Deleting user with ID: %s", userId)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User not found for deletion: %s", userId)
			return NewUserNotFoundError(userId)
		}
		s.log.Error("Error checking if user exists for deletion: %v", err)
		return NewDBError("checking if user exists for deletion", err.Error())
	}
	switch user.Status {
	case model.Deleted:
		s.log.Error("User is already deleted: %s", userId)
		return NewUserMarkedForDeletionError(userId)
	case model.Inactive:
		s.log.Error("User is inactive and cannot be deleted: %s", userId)
		return NewUserInactiveError(userId)
	case model.Banned:
		s.log.Error("User is banned and cannot be deleted: %s", userId)
		return NewUserBannedError(userId, "User violated terms and conditions of the platform")
	}

	_, err = s.userQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"userId": userId},
		bson.M{"$set": bson.M{
			"status":    model.Deleted,
			"deletedAt": primitive.NewDateTimeFromTime(time.Now()),
			"updatedAt": primitive.NewDateTimeFromTime(time.Now()),
		}},
		nil,
	)
	if err != nil {
		s.log.Error("Error deleting user: %v", err)
		return NewDBError("deleting user", err.Error())
	}

	s.log.Debug("User marked for deletion successfully: %s", userId)
	return nil
}

func (s *userService) ChangePassword(userId string, oldPassword string, newPassword string) network.ApiError {
	s.log.Debug("Changing password for user ID: %s", userId)
	user, err := s.userQueryBuilder.SingleQuery().FilterOne(bson.M{"userId": userId}, nil)
	if err != nil {
		if mongo.IsNoDocumentFoundError(err) {
			s.log.Error("User not found for password change: %s", userId)
			return NewUserNotFoundError(userId)
		}
		s.log.Error("Error checking if user exists for password change: %v", err)
		return NewDBError("checking if user exists for password change", err.Error())
	}

	switch user.Status {
	case model.Deleted:
		s.log.Error("User is deleted and cannot change password: %s", userId)
		return NewUserDeletedError(userId)
	case model.Inactive:
		s.log.Error("User is inactive and cannot change password: %s", userId)
		return NewUserInactiveError(userId)
	case model.Banned:
		s.log.Error("User is banned and cannot change password: %s", userId)
		return NewUserBannedError(userId, "User violated terms and conditions of the platform")
	}
	hasntSetPassword := user.PasswordHash == EMPTY_PASSWORD_HASH
	var newPasswordHash string
	if hasntSetPassword {
		s.log.Error("User has not set a password yet: %s", userId)
		newPasswordHash, err = utils.HashPassword(newPassword)
		if err != nil {
			s.log.Error("Error hashing new password: %v", err)
			return NewDBError("hashing new password", err.Error())
		}
	} else {
		s.log.Debug("Validating old password for user ID: %s", userId)
		err = s.ValidateUserPassword(user, oldPassword)
		if err != nil {
			s.log.Error("Invalid old password for user ID: %s", userId)
			return NewWrongOldPasswordError(userId)
		}
		s.log.Debug("Old password validated successfully for user ID: %s", userId)
		newPasswordHash, err = utils.HashPassword(newPassword)
		if err != nil {
			s.log.Error("Error hashing new password: %v", err)
			return NewDBError("hashing new password", err.Error())
		}
	}
	_, err = s.userQueryBuilder.SingleQuery().UpdateOne(
		bson.M{"userId": userId},
		bson.M{"$set": bson.M{"passwordHash": newPasswordHash, "updatedAt": primitive.NewDateTimeFromTime(time.Now())}},
		nil,
	)
	if err != nil {
		s.log.Error("Error setting new password for user: %v", err)
		return NewDBError("setting new password for user", err.Error())
	}
	s.log.Debug("Password changed successfully for user ID: %s", userId)
	return nil
}

func (s *userService) JoinCommunity(userId string, communityId string) network.ApiError {
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

func (s *userService) LeaveCommunity(userId string, communityId string) network.ApiError {
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

func (s *userService) SearchUsers(userId string, query string, page int, limit int) ([]*model.SearchUser, network.ApiError) {
	s.log.Debug("Searching users with query: %s, page: %d, limit: %d", query, page, limit)
	regexPattern := primitive.Regex{Pattern: query, Options: "i"}

	aggregationPipeline := s.searchUsersAggregator.SingleAggregate()
	aggregationPipeline.AllowDiskUse(true)

	// Match documents that contain the query in username, fullName, or bio
	// And exclude the current user and deleted/banned users
	aggregationPipeline.Match(bson.M{
		"userId": bson.M{"$ne": userId},
		"status": bson.M{"$nin": []model.UserStatus{model.Deleted, model.Banned}},
		"$or": []bson.M{
			{"username": bson.M{"$regex": regexPattern}},
			{"fullName": bson.M{"$regex": regexPattern}},
			{"bio": bson.M{"$regex": regexPattern}},
		},
	})

	// Project only the fields needed for SearchUser model
	aggregationPipeline.Project(bson.M{
		"userId":      1,
		"username":    1,
		"email":       1,
		"avatar":      "$avatar.profile.url",
		"background":  "$avatar.background.url",
		"followers":   1,
		"follows":     1,
		"status":      1,
		"isFollowing": bson.M{"$in": []any{userId, bson.M{"$ifNull": []any{"$followers", []string{}}}}},
		"isFollowed":  bson.M{"$in": []any{userId, bson.M{"$ifNull": []any{"$follows", []string{}}}}},
		"isBlocked":   bson.M{"$in": []any{userId, bson.M{"$ifNull": []any{"$preferences.blockList", []string{}}}}},
		"_id":         0,
	})

	// Execute the aggregation
	results, err := aggregationPipeline.ExecPaginated(int64(page), int64(limit))
	if err != nil {
		s.log.Error("Error searching users: %v", err)
		return nil, NewDBError("searching users", err.Error())
	}

	s.log.Debug("Found %d users matching query: %s", len(results), query)
	return results, nil
}

func (s *userService) FollowUser(userId string, followUserId string) network.ApiError {
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

func (s *userService) UnfollowUser(userId string, unfollowUserId string) network.ApiError {
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

func (s *userService) BlockUser(userId string, blockUserId string) network.ApiError {
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

func (s *userService) UnblockUser(userId string, unblockUserId string) network.ApiError {
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

func (s *userService) AddModerator(userId string, communityId string) network.ApiError {
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
					"moderatedWavelengths": communityId,
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

func (s *userService) RemoveModerator(userId string, communityId string) network.ApiError {
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
					"moderatedWavelengths": communityId,
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
