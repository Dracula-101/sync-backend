package application

import (
	"fmt"

	"github.com/spf13/cobra"

	"sync-backend/internal/api/middleware"
	"sync-backend/internal/api/route"
	"sync-backend/internal/infrastructure/config"
	"sync-backend/internal/server"
	"sync-backend/pkg/console"
	"sync-backend/pkg/logger"
)

// ServeCommand test command
type ServeCommand struct{}

func (s *ServeCommand) Short() string {
	return "serve application"
}

func (s *ServeCommand) Setup(cmd *cobra.Command) {
}

func (s *ServeCommand) Run() console.CommandRunner {
	return func(
		logger logger.Logger,
		server server.Server,
		middlewares middleware.Middlewares,
		baseRateLimiter middleware.BaseRateLimitMiddleware,
		routes route.Routes,
		appConfig *config.Config,
	) {
		logger.Info("Loading environment variables")
		config.LoadEnv()
		logger.Info("Starting server")
		middlewares.Setup()
		baseApiPrefix := fmt.Sprintf("%s/v%s", appConfig.API.Prefix, appConfig.API.Version)
		apiRouterGroup := server.Group(baseApiPrefix, baseRateLimiter.Handle())
		routes.Setup(apiRouterGroup)

		if err := server.Run(); err != nil {
			logger.Fatal(err)
		}
	}
}

func NewServeCommand() *ServeCommand {
	return &ServeCommand{}
}
