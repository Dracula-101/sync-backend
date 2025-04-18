package application

import (
	"context"

	"sync-backend/api/auth"
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
}

func (m *appModule) GetInstance() *appModule {
	return m
}

func (m *appModule) Controllers() []network.Controller {
	return []network.Controller{
		auth.NewAuthController(m.AuthService),
	}
}

func (m *appModule) RootMiddlewares() []network.RootMiddleware {
	return []network.RootMiddleware{
		coreMW.NewErrorCatcher(), // NOTE: this should be the first handler to be mounted
		// authMW.NewKeyProtection(m.AuthService),
		coreMW.NewNotFound(),
	}
}
func NewAppModule(context context.Context, env *config.Env, config *config.Config, db mongo.Database, store redis.Store) Module {
	authService := auth.NewAuthService(db, config)

	return &appModule{
		Context:     context,
		Env:         env,
		Config:      config,
		DB:          db,
		Store:       store,
		AuthService: authService,
	}
}
