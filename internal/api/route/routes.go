package route

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module exports dependency to container
var Module = fx.Options(
	fx.Provide(NewRouter),
	fx.Provide(NewAuthRoute),
)

type Routes []Route

// Route interface
type Route interface {
	Setup(apiRouteGroup *gin.RouterGroup)
}

// NewRoutes sets up routes
func NewRouter(
	authRoutes *AuthRoutes,
) Routes {
	return Routes{
		authRoutes,
	}
}

// Setup all the route
func (r Routes) Setup(apiRouteGroup *gin.RouterGroup) {
	for _, route := range r {
		route.Setup(apiRouteGroup)
	}
}
