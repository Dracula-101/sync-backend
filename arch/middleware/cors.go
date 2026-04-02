package middleware

import (
	"fmt"
	"strings"
	"sync-backend/arch/config"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type cors struct {
	network.BaseMiddleware
	logger utils.AppLogger
	config config.CORSConfig
}

func NewCORS(config config.CORSConfig) network.RootMiddleware {
	return &cors{
		BaseMiddleware: network.NewBaseMiddleware(),
		logger:         utils.NewServiceLogger("CORS"),
		config:         config,
	}
}

func (m *cors) Attach(engine *gin.Engine) {
	if m.config.Enabled {
		engine.Use(m.Handler)
		m.logger.Info("CORS middleware enabled")
	} else {
		m.logger.Info("CORS middleware disabled")
	}
}

func (m *cors) Handler(ctx *gin.Context) {
	origin := ctx.Request.Header.Get("Origin")

	// Set CORS headers
	if m.config.AllowOrigin == "*" {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	} else if origin != "" {
		// Check if origin is in allowed origins
		allowedOrigins := strings.Split(m.config.AllowOrigin, ",")
		for _, allowedOrigin := range allowedOrigins {
			if strings.TrimSpace(allowedOrigin) == origin {
				ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
	}

	if m.config.AllowCredentials {
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if m.config.AllowMethods != "" {
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", m.config.AllowMethods)
	}

	if m.config.AllowHeaders != "" {
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", m.config.AllowHeaders)
	}

	if m.config.ExposeHeaders != "" {
		ctx.Writer.Header().Set("Access-Control-Expose-Headers", m.config.ExposeHeaders)
	}

	if m.config.MaxAge > 0 {
		ctx.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", m.config.MaxAge))
	}

	// Handle preflight OPTIONS requests
	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(204)
		return
	}

	ctx.Next()
}
