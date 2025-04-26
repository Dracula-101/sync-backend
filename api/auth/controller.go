package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/api/location"
	"sync-backend/api/user"
	"sync-backend/arch/common"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type authController struct {
	logger utils.AppLogger
	network.BaseController
	common.ContextPayload
	authProvider    network.AuthenticationProvider
	authService     AuthService
	userService     user.UserService
	locationService location.LocationService
}

func NewAuthController(
	authService AuthService,
	authProvider network.AuthenticationProvider,
	userService user.UserService,
	locationService location.LocationService,
) network.Controller {
	return &authController{
		logger:          utils.NewServiceLogger("AuthController"),
		BaseController:  network.NewBaseController("/api/v1/auth", authProvider),
		ContextPayload:  common.NewContextPayload(),
		authProvider:    authProvider,
		authService:     authService,
		userService:     userService,
		locationService: locationService,
	}
}

func (c *authController) MountRoutes(group *gin.RouterGroup) {
	c.logger.Info("Mounting auth routes")
	group.POST("/signup", c.SignUp)
	group.POST("/login", c.Login)
	group.POST("/google", c.GoogleLogin)
	group.POST("/logout", c.authProvider.Middleware(), c.Logout)
}

func (c *authController) SignUp(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewSignUpRequest())
	if err != nil {
		return
	}
	exists, err := c.userService.FindUserByEmail(body.Email)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	if exists != nil {
		c.Send(ctx).ConflictError("User with this email already exists", nil)
		return
	}

	c.SetRequestDetails(ctx, &body.BaseRequest)
	data, err := c.authService.SignUp(body)

	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User created successfully", data)
}

func (c *authController) Login(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewLoginRequest())
	if err != nil {
		return
	}
	c.SetRequestDetails(ctx, &body.BaseRequest)
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
	c.SetRequestDetails(ctx, &body.BaseRequest)
	data, err := c.authService.GoogleLogin(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessDataResponse("User logged in with google successfully", data)
}

func (c *authController) Logout(ctx *gin.Context) {
	body, err := network.ReqBody(ctx, dto.NewLogoutRequest())
	if err != nil {
		return
	}
	c.SetRequestDetails(ctx, &body.BaseRequest)
	err = c.authService.Logout(body)
	if err != nil {
		c.Send(ctx).MixedError(err)
		return
	}
	c.Send(ctx).SuccessMsgResponse("User logged out successfully")
}
