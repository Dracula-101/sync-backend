package application

import (
	"context"

	"sync-backend/api/auth"
	authMW "sync-backend/api/auth/middleware"
	"sync-backend/api/session"
	"sync-backend/api/token"
	"sync-backend/api/user"
	"sync-backend/arch/config"
	coreMW "sync-backend/arch/middleware"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/arch/redis"
	"sync-backend/utils"
)

type Module network.Module[appModule]

type appModule struct {
	Context        context.Context
	Logger         utils.AppLogger
	Env            *config.Env
	Config         *config.Config
	DB             mongo.Database
	Store          redis.Store
	AuthService    auth.AuthService
	UserService    user.UserService
	SessionService session.SessionService
	TokenService   token.TokenService
}

func (m *appModule) GetInstance() *appModule {
	return m
}

func (m *appModule) Controllers() []network.Controller {
	return []network.Controller{
		auth.NewAuthController(m.Logger, m.AuthService, m.UserService, m.AuthenticationProvider()),
	}
}

func (m *appModule) AuthenticationProvider() network.AuthenticationProvider {
	return authMW.NewAuthenticationProvider(m.TokenService, m.UserService)
}

func (m *appModule) RootMiddlewares() []network.RootMiddleware {
	middlewares := []network.RootMiddleware{}
	middlewares = append(middlewares, coreMW.NewErrorCatcher(&m.Logger))
	if m.Config.API.RateLimit.Enabled {
		middlewares = append(middlewares, coreMW.NewRateLimiter(m.Store, *m.Config))
	}

	return middlewares
}
func NewAppModule(context context.Context, logger utils.AppLogger, env *config.Env, config *config.Config, db mongo.Database, store redis.Store) Module {
	tokenService := token.NewTokenService(config)
	sessionService := session.NewSessionService(db)
	userService := user.NewUserService(db)
	authService := auth.NewAuthService(userService, sessionService, tokenService, config)
	return &appModule{
		Context:        context,
		Logger:         logger,
		Env:            env,
		Config:         config,
		DB:             db,
		Store:          store,
		AuthService:    authService,
		UserService:    userService,
		SessionService: sessionService,
		TokenService:   tokenService,
	}
}
