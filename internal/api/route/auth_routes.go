package route

import (
	"sync-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	logger logger.Logger
}

func NewAuthRoute(logger logger.Logger) *AuthRoutes {
	return &AuthRoutes{
		logger: logger,
	}
}

func (a *AuthRoutes) Setup(apiRouteGroup *gin.RouterGroup) {
	a.logger.Info("Setting up auth routes")
	authGroup := apiRouteGroup.Group("/auth")
	authGroup.POST("/login", a.Login)
	authGroup.POST("/register", a.Register)
	authGroup.POST("/refresh-token", a.RefreshToken)
	authGroup.GET("/logout", a.Logout)
}

func (a *AuthRoutes) Login(c *gin.Context) {
	// Handle login logic here
	c.JSON(200, gin.H{"message": "Login successful"})
}

func (a *AuthRoutes) Register(c *gin.Context) {
	// Handle register logic here
	c.JSON(200, gin.H{"message": "Registration successful"})
}

func (a *AuthRoutes) RefreshToken(c *gin.Context) {
	// Handle refresh token logic here
	c.JSON(200, gin.H{"message": "Token refreshed successfully"})
}

func (a *AuthRoutes) Logout(c *gin.Context) {
	// Handle logout logic here
	c.JSON(200, gin.H{"message": "Logout successful"})
}
