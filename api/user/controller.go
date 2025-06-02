package user

import (
	"fmt"
	"sync-backend/api/common/location"
	"sync-backend/api/user/dto"
	"sync-backend/arch/common"
	coreMW "sync-backend/arch/middleware"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type userController struct {
	logger utils.AppLogger
	network.BaseController
	common.ContextPayload
	authProvider    network.AuthenticationProvider
	uploadProvider  coreMW.UploadProvider
	userService     UserService
	locationService location.LocationService
}

func NewUserController(
	authProvider network.AuthenticationProvider,
	uploadProvider coreMW.UploadProvider,
	userService UserService,
	locationService location.LocationService,
) network.Controller {
	return &userController{
		logger:          utils.NewServiceLogger("UserController"),
		BaseController:  network.NewBaseController("/user", nil),
		ContextPayload:  common.NewContextPayload(),
		authProvider:    authProvider,
		uploadProvider:  uploadProvider,
		userService:     userService,
		locationService: locationService,
	}
}

func (c *userController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting user routes")
	group.Use(c.authProvider.Middleware())
	group.GET("/me", c.GetMe)
	group.PUT("me", c.uploadProvider.Middleware("avatar", "background"), c.UpdateMe)
	group.DELETE("/me", c.DeleteMe)
	group.PUT("/me/preferences", c.UpdatePreferences)
	group.GET("/me/password", c.ChangePassword)

	group.GET("/search", c.SearchUsers)

	group.GET("/:userId", c.GetUserById)
	group.POST("/follow/:userId", c.FollowUser)

	group.POST("/unfollow/:userId", c.UnfollowUser)
	group.POST("/block/:userId", c.BlockUser)
	group.POST("/unblock/:userId", c.UnblockUser)

}

func (c *userController) GetMe(ctx *gin.Context) {

	userId := c.ContextPayload.MustGetUserId(ctx)
	user, err := c.userService.FindUserById(*userId)

	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if user == nil {
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", *userId),
			nil,
		)
		return
	}

	c.Send(ctx).SuccessDataResponse("Profile fetched successfully", user)
}

func (c *userController) UpdateMe(ctx *gin.Context) {
	userId := c.ContextPayload.MustGetUserId(ctx)
	form, err := network.ReqForm(ctx, dto.NewUpdateUserRequest())
	if err != nil {
		return
	}

	avatarFile, err := c.uploadProvider.GetUploadedFiles(ctx, "avatar").First()
	if err == nil {
		form.AvatarFilePath = &avatarFile.Path
	}
	backgroundFile, err := c.uploadProvider.GetUploadedFiles(ctx, "background").First()
	if err == nil {
		form.BackgroundFilePath = &backgroundFile.Path
	}

	user, err := c.userService.UpdateUserProfile(*userId, &form.Bio, form.AvatarFilePath, form.BackgroundFilePath)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	if user == nil {
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", *userId),
			nil,
		)
		return
	}

	c.Send(ctx).SuccessDataResponse("Profile updated successfully", user)
}

func (c *userController) DeleteMe(ctx *gin.Context) {
	userId := c.ContextPayload.MustGetUserId(ctx)

	err := c.userService.DeleteUser(*userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("User marked for deletion successfully")
}

func (c *userController) UpdatePreferences(ctx *gin.Context) {
	userId := c.ContextPayload.MustGetUserId(ctx)
	form, err := network.ReqForm(ctx, dto.NewUpdateUserPreferencesRequest())
	if err != nil {
		return
	}

	user, err := c.userService.FindUserById(*userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	if user == nil {
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", *userId),
			nil,
		)
		return
	}

	if form.Language != nil {
		user.Preferences.Language = common.GetLanguageByID(*form.Language).ToDetail()
	}
	if form.Theme != nil {
		user.Preferences.Theme = *form.Theme
	}
	if form.Timezone != nil {
		user.Preferences.Timezone = common.GetTimeZone(*form.Timezone).ToDetail()
	}
	if form.ShowEmailNotifications != nil {
		user.Preferences.Notifications.Email = *form.ShowEmailNotifications
	}
	if form.ShowMobileNotifications != nil {
		user.Preferences.Notifications.Push = *form.ShowMobileNotifications
	}
	if form.ShowSensitiveContent != nil {
		user.Preferences.ContentSettings.ShowSensitiveContent = *form.ShowSensitiveContent
	}
	if form.ShowAdultContent != nil {
		user.Preferences.ContentSettings.ShowAdultContent = *form.ShowAdultContent
	}
	if form.IsProfileVisible != nil {
		user.Preferences.PrivacySettings.IsProfileVisible = *form.IsProfileVisible
	}
	if form.IsEmailVisible != nil {
		user.Preferences.PrivacySettings.IsEmailVisible = *form.IsEmailVisible
	}
	if form.IsJoinedWavelengthsVisible != nil {
		user.Preferences.PrivacySettings.IsJoinedWavelengthsVisible = *form.IsJoinedWavelengthsVisible
	}
	if form.FollowersVisible != nil {
		user.Preferences.PrivacySettings.FollowersVisible = *form.FollowersVisible
	}

	_, err = c.userService.UpdateUserPreferences(*userId, user.Preferences)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Preferences updated successfully")
}

func (c *userController) ChangePassword(ctx *gin.Context) {
	userId := c.ContextPayload.MustGetUserId(ctx)
	form, err := network.ReqForm(ctx, dto.NewChangePasswordRequest())
	if err != nil {
		return
	}

	err = c.userService.ChangePassword(*userId, form.OldPassword, form.NewPassword)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Password changed successfully")
}

func (c *userController) SearchUsers(ctx *gin.Context) {
	userId := c.ContextPayload.MustGetUserId(ctx)
	params, err := network.ReqQuery(ctx, dto.NewSearchUsersRequest())
	if err != nil {
		return
	}

	users, err := c.userService.SearchUsers(*userId, params.Query, params.Page, params.Limit)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessDataResponse("Users fetched successfully", users)
}

func (c *userController) GetUserById(ctx *gin.Context) {
	params, err := network.ReqParams(ctx, dto.NewGetUserRequest())
	if err != nil {
		return
	}
	userId := params.UserId
	user, err := c.userService.FindUserById(userId)

	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	if user == nil {
		c.Send(ctx).NotFoundError(
			"User not found",
			fmt.Sprintf("User with ID %s not found", userId),
			nil,
		)
		return
	}

	c.Send(ctx).SuccessDataResponse("Profile fetched successfully", user)
}

func (c *userController) FollowUser(ctx *gin.Context) {
	followUserId := ctx.Param("userId")
	userId := c.ContextPayload.MustGetUserId(ctx)

	if *userId == followUserId {
		c.Send(ctx).MixedError(NewSelfActionError("follow"))
		return
	}

	err := c.userService.FollowUser(*userId, followUserId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Followed user successfully")
}

func (c *userController) UnfollowUser(ctx *gin.Context) {
	unfollowUserId := ctx.Param("userId")
	userId := c.ContextPayload.MustGetUserId(ctx)

	if *userId == unfollowUserId {
		c.Send(ctx).MixedError(NewSelfActionError("unfollow"))
		return
	}

	err := c.userService.UnfollowUser(*userId, unfollowUserId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Unfollowed user successfully")
}

func (c *userController) BlockUser(ctx *gin.Context) {
	blockUserId := ctx.Param("userId")
	userId := c.ContextPayload.MustGetUserId(ctx)

	if *userId == blockUserId {
		c.Send(ctx).MixedError(NewSelfActionError("block"))
		return
	}

	err := c.userService.BlockUser(*userId, blockUserId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Blocked user successfully")
}

func (c *userController) UnblockUser(ctx *gin.Context) {
	unblockUserId := ctx.Param("userId")
	userId := c.ContextPayload.MustGetUserId(ctx)

	if *userId == unblockUserId {
		c.Send(ctx).MixedError(NewSelfActionError("unblock"))
		return
	}

	err := c.userService.UnblockUser(*userId, unblockUserId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}

	c.Send(ctx).SuccessMsgResponse("Unblocked user successfully")
}
