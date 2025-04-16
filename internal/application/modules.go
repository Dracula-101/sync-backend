package application

import (
	"sync-backend/internal/api/middleware"
	"sync-backend/internal/api/route"
	"sync-backend/internal/infrastructure/config"
	"sync-backend/internal/server"
	"sync-backend/pkg/logger"

	"go.uber.org/fx"
)

// Module exports all dependencies for fx app
var CommonModules = fx.Options(
	fx.Provide(logger.GetLogger),
	fx.Provide(config.GetConfig),
	fx.Provide(server.NewServer),
	route.Module,
	middleware.Module,
)
