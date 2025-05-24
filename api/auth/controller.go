package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/api/common/location"
	"sync-backend/api/user"
	"sync-backend/arch/common"
	coreMW "sync-backend/arch/middleware"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type authController struct {
	logger utils.AppLogger
	network.BaseController
	common.ContextPayload
	authProvider     network.AuthenticationProvider
	uploadProvider   coreMW.UploadProvider
	locationProvider network.LocationProvider
	authService      AuthService
	userService      user.UserService
	locationService  location.LocationService
}

func NewAuthController(
	authProvider network.AuthenticationProvider,
	locationProvider network.LocationProvider,
	uploadProvider coreMW.UploadProvider,
	authService AuthService,
	userService user.UserService,
	locationService location.LocationService,
) network.Controller {
	return &authController{
		logger:           utils.NewServiceLogger("AuthController"),
		BaseController:   network.NewBaseController("/auth", authProvider),
		ContextPayload:   common.NewContextPayload(),
		authProvider:     authProvider,
		uploadProvider:   uploadProvider,
		locationProvider: locationProvider,
		authService:      authService,
		userService:      userService,
		locationService:  locationService,
	}
}

func (c *authController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting auth routes")
	group.POST("/signup", c.locationProvider.Middleware(), c.uploadProvider.Middleware("profile_photo", "background_photo"), c.SignUp)
	group.POST("/login", c.locationProvider.Middleware(), c.Login)
	group.POST("/google", c.locationProvider.Middleware(), c.GoogleLogin)
	group.POST("/logout", c.authProvider.Middleware(), c.Logout)
	group.POST("/forgot-password", c.ForgotPassword)
	group.POST("/refresh-token", c.locationProvider.Middleware(), c.RefreshToken)
}

func (c *authController) SignUp(ctx *gin.Context) {
	body, err := network.ReqForm(ctx, dto.NewSignUpRequest())
	if err != nil {
		return
	}
	exists, err := c.userService.FindUserByEmail(body.Email)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	if exists != nil {
		c.Send(ctx).MixedError(NewUserExistsByEmailError(body.Email))
		return
	}

	exists, err = c.userService.FindUserByUsername(body.UserName)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	if exists != nil {
		c.Send(ctx).MixedError(NewUserExistsByUsernameError(body.UserName))
		return
	}

	profile := c.uploadProvider.GetUploadedFiles(ctx, "profile_photo")
	backgroundPic := c.uploadProvider.GetUploadedFiles(ctx, "background_photo")
	if len(profile.Files) != 0 {
		body.ProfileFilePath = profile.Files[0].Path
	}
	if len(backgroundPic.Files) != 0 {
		body.BackgroundFilePath = backgroundPic.Files[0].Path
	}

	c.SetRequestDeviceDetails(ctx, &body.BaseDeviceRequest)
	c.SetRequestLocationDetails(ctx, &body.BaseLocationRequest)
	data, err := c.authService.SignUp(body)

	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User created successfully", data)
	c.uploadProvider.DeleteUploadedFiles(ctx, "profile_photo")
	c.uploadProvider.DeleteUploadedFiles(ctx, "background_photo")
}

func (c *authController) Login(ctx *gin.Context) {
	body, err := network.ReqForm(ctx, dto.NewLoginRequest())
	if err != nil {
		return
	}
	c.SetRequestDeviceDetails(ctx, &body.BaseDeviceRequest)
	c.SetRequestLocationDetails(ctx, &body.BaseLocationRequest)
	data, err := c.authService.Login(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User logged in successfully", data)
}

func (c *authController) GoogleLogin(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewGoogleLoginRequest())
	if err != nil {
		return
	}
	c.SetRequestDeviceDetails(ctx, &body.BaseDeviceRequest)
	c.SetRequestLocationDetails(ctx, &body.BaseLocationRequest)
	data, err := c.authService.GoogleLogin(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User logged in with google successfully", data)
}

func (c *authController) Logout(ctx *gin.Context) {
	userId := *c.MustGetUserId(ctx)
	err := c.authService.Logout(userId)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("User logged out successfully")
}

func (c *authController) ForgotPassword(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewForgotPassRequest())
	if err != nil {
		return
	}

	c.SetRequestDeviceDetails(ctx, &body.BaseDeviceRequest)
	err = c.authService.ForgotPassword(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("Password reset link sent to your email")
}

func (c *authController) RefreshToken(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewRefreshTokenRequest())
	if err != nil {
		return
	}
	c.SetRequestDeviceDetails(ctx, &body.BaseDeviceRequest)
	c.SetRequestLocationDetails(ctx, &body.BaseLocationRequest)
	data, err := c.authService.RefreshToken(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("Token refreshed successfully", data)
}
