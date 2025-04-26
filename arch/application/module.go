package application

import (
	"context"

	"sync-backend/api/auth"
	authMW "sync-backend/api/auth/middleware"
	"sync-backend/api/location"
	"sync-backend/api/session"
	"sync-backend/api/token"
	"sync-backend/api/user"
	"sync-backend/arch/config"
	coreMW "sync-backend/arch/middleware"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	pg "sync-backend/arch/postgres"
	"sync-backend/arch/redis"
)

type Module network.Module[appModule]

type appModule struct {
	Context         context.Context
	Env             *config.Env
	Config          *config.Config
	DB              mongo.Database
	IpDB            pg.Database
	Store           redis.Store
	AuthService     auth.AuthService
	UserService     user.UserService
	SessionService  session.SessionService
	LocationService location.LocationService
	TokenService    token.TokenService
}

func (m *appModule) GetInstance() *appModule {
	return m
}

func (m *appModule) Controllers() []network.Controller {
	return []network.Controller{
		auth.NewAuthController(m.AuthService, m.AuthenticationProvider(), m.UserService, m.LocationService),
	}
}

func (m *appModule) AuthenticationProvider() network.AuthenticationProvider {
	return authMW.NewAuthenticationProvider(m.TokenService, m.UserService)
}

func (m *appModule) RootMiddlewares() []network.RootMiddleware {
	middlewares := []network.RootMiddleware{}
	middlewares = append(middlewares, coreMW.NewErrorCatcher())
	if m.Config.API.RateLimit.Enabled {
		middlewares = append(middlewares, coreMW.NewRateLimiter(m.Store, *m.Config))
	}

	return middlewares
}
func NewAppModule(context context.Context, env *config.Env, config *config.Config, db mongo.Database, ipDb pg.Database, store redis.Store) Module {
	locationService := location.NewLocationService(ipDb)
	tokenService := token.NewTokenService(config)
	sessionService := session.NewSessionService(db)
	userService := user.NewUserService(db)
	authService := auth.NewAuthService(userService, sessionService, locationService, tokenService, config)
	return &appModule{
		Context:         context,
		Env:             env,
		Config:          config,
		DB:              db,
		IpDB:            ipDb,
		Store:           store,
		AuthService:     authService,
		UserService:     userService,
		LocationService: locationService,
		SessionService:  sessionService,
		TokenService:    tokenService,
	}
}
