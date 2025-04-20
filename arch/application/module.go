package application

import (
	"context"

	"sync-backend/api/auth"
	authMW "sync-backend/api/auth/middleware"
	"sync-backend/api/user"
	"sync-backend/arch/config"
	coreMW "sync-backend/arch/middleware"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/arch/redis"
)

type Module network.Module[appModule]

type appModule struct {
	Context     context.Context
	Env         *config.Env
	Config      *config.Config
	DB          mongo.Database
	Store       redis.Store
	AuthService auth.AuthService
	UserService user.UserService
}

func (m *appModule) GetInstance() *appModule {
	return m
}

func (m *appModule) Controllers() []network.Controller {
	return []network.Controller{
		auth.NewAuthController(m.AuthService, m.AuthenticationProvider()),
	}
}

func (m *appModule) AuthenticationProvider() network.AuthenticationProvider {
	return authMW.NewAuthenticationProvider(m.AuthService, m.UserService)
}

func (m *appModule) RootMiddlewares() []network.RootMiddleware {
	middlewares := []network.RootMiddleware{}
	middlewares = append(middlewares, coreMW.NewErrorCatcher())
	middlewares = append(middlewares, coreMW.NewNotFound())
	middlewares = append(middlewares, coreMW.NewMethodNotAllowed())
	if m.Config.API.RateLimit.Enabled {
		middlewares = append(middlewares, coreMW.NewRateLimiter(m.Store, *m.Config))
	}

	return middlewares
}
func NewAppModule(context context.Context, env *config.Env, config *config.Config, db mongo.Database, store redis.Store) Module {
	userService := user.NewUserService(db)
	authService := auth.NewAuthService(db, userService, config)
	return &appModule{
		Context:     context,
		Env:         env,
		Config:      config,
		DB:          db,
		Store:       store,
		AuthService: authService,
		UserService: userService,
	}
}
