package application

import (
	"context"

	"sync-backend/api/auth"
	authMW "sync-backend/api/auth/middleware"
	"sync-backend/api/comment"
	"sync-backend/api/common/location"
	"sync-backend/api/common/media"
	"sync-backend/api/common/session"
	"sync-backend/api/common/token"
	"sync-backend/api/community"
	"sync-backend/api/post"
	"sync-backend/api/system"
	"sync-backend/api/user"
	"sync-backend/arch/config"
	coreMW "sync-backend/arch/middleware"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	pg "sync-backend/arch/postgres"
	"sync-backend/arch/redis"

	"github.com/gin-gonic/gin"
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
	MediaService    media.MediaService

	// Services
	AuthService      auth.AuthService
	CommunityService community.CommunityService
	PostService      post.PostService
	CommentService   comment.CommentService
	SystemService    system.SystemService
}

func (m *appModule) GetInstance() *appModule {
	return m
}

func (m *appModule) Controllers() []network.Controller {
	return []network.Controller{
		auth.NewAuthController(m.AuthenticationProvider(), m.LocationProvider(), m.UploadProvider(), m.AuthService, m.UserService, m.LocationService),
		community.NewCommunityController(m.AuthenticationProvider(), m.UploadProvider(), m.CommunityService),
		user.NewUserController(m.AuthenticationProvider(), m.UploadProvider(), m.UserService, m.CommunityService, m.LocationService),
		post.NewPostController(m.AuthenticationProvider(), m.UploadProvider(), m.PostService),
		comment.NewCommentController(m.AuthenticationProvider(), m.LocationProvider(), m.CommentService),
		system.NewSystemController(m.SystemService),
	}
}

func (m *appModule) AuthenticationProvider() network.AuthenticationProvider {
	return authMW.NewAuthenticationProvider(m.TokenService, m.UserService, m.SessionService, m.Store)
}

func (m *appModule) LocationProvider() network.LocationProvider {
	return authMW.NewLocationProvider(m.LocationService, m.Store)
}

func (m *appModule) UploadProvider() coreMW.UploadProvider {
	return coreMW.NewUploadProvider()
}

func (m *appModule) RootMiddlewares() []network.RootMiddleware {
	middlewares := []network.RootMiddleware{}
	middlewares = append(middlewares, coreMW.NewErrorCatcher())
	if m.Config.API.RateLimit.Enabled {
		middlewares = append(middlewares, coreMW.NewRateLimiter(m.Store, *m.Config))
	}

	return middlewares
}

func NewAppModule(context context.Context, env *config.Env, config *config.Config, db mongo.Database, ipDb pg.Database, store redis.Store, engine *gin.Engine) Module {
	mediaService := media.NewMediaService(*env)
	locationService := location.NewLocationService(ipDb)
	tokenService := token.NewTokenService(config)
	sessionService := session.NewSessionService(db)
	systemService := system.NewSystemService(config, db, store, engine)

	userService := user.NewUserService(db, mediaService)
	authService := auth.NewAuthService(config, userService, sessionService, tokenService)
	communityService := community.NewCommunityService(db, mediaService)
	postService := post.NewPostService(db, userService, communityService, mediaService)
	commentService := comment.NewCommentService(db)

	return &appModule{
		Context: context,
		Env:     env,
		Config:  config,
		DB:      db,
		IpDB:    ipDb,
		Store:   store,

		// Common services
		UserService:     userService,
		LocationService: locationService,
		SessionService:  sessionService,
		TokenService:    tokenService,
		MediaService:    mediaService,
		SystemService:   systemService,

		// Services
		AuthService:      authService,
		CommunityService: communityService,
		PostService:      postService,
		CommentService:   commentService,
	}
}
