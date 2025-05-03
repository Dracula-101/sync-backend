package application

import (
	"context"

	"sync-backend/api/auth"
	authMW "sync-backend/api/auth/middleware"
	"sync-backend/api/common/location"
	"sync-backend/api/common/session"
	"sync-backend/api/common/token"
	"sync-backend/api/community"
	"sync-backend/api/post"
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
	Context context.Context
	Env     *config.Env
	Config  *config.Config
	DB      mongo.Database
	IpDB    pg.Database
	Store   redis.Store

	// Common services
	UserService     user.UserService
	SessionService  session.SessionService
	LocationService location.LocationService
	TokenService    token.TokenService

	// Services
	AuthService      auth.AuthService
	CommunityService community.CommunityService
	PostService      post.PostService
}

func (m *appModule) GetInstance() *appModule {
	return m
}

func (m *appModule) Controllers() []network.Controller {
	return []network.Controller{
		auth.NewAuthController(m.AuthService, m.AuthenticationProvider(), m.UserService, m.LocationService),
		community.NewCommunityController(m.CommunityService, m.AuthenticationProvider()),
		user.NewUserController(m.AuthenticationProvider(), m.UserService, m.LocationService),
		post.NewPostController(m.PostService, m.AuthenticationProvider()),
	}
}

func (m *appModule) AuthenticationProvider() network.AuthenticationProvider {
	return authMW.NewAuthenticationProvider(m.TokenService, m.UserService, m.SessionService, m.Store)
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
	sessionService := session.NewSessionService(db, locationService)

	userService := user.NewUserService(db)
	authService := auth.NewAuthService(config, userService, sessionService, locationService, tokenService)
	communityService := community.NewCommunityService(db)
	postService := post.NewPostService(db, userService, communityService)
	return &appModule{
		Context:         context,
		Env:             env,
		Config:          config,
		DB:              db,
		IpDB:            ipDb,
		Store:           store,
		UserService:     userService,
		LocationService: locationService,
		SessionService:  sessionService,
		TokenService:    tokenService,

		// Services
		AuthService:      authService,
		CommunityService: communityService,
		PostService:      postService,
	}
}
