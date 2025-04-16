package server

import (
	"net/http"
	"strings"
	"sync-backend/internal/infrastructure/config"
	"sync-backend/pkg/logger"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
)

// ServerConfig holds the configuration for the server
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
	MaxConnections int           `mapstructure:"max_connections"`
	MaxIdleConns   int           `mapstructure:"max_idle_connections"`
	RouterConfig   RouterConfig  `mapstructure:"router"`
}

type RouterConfig struct {
	Prefix    string          `mapstructure:"prefix"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	CORS      CORSConfig      `mapstructure:"cors"`
}

type RateLimitConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	MaxRequests int    `mapstructure:"max_requests"`
	Burst       int    `mapstructure:"burst"`
	Window      string `mapstructure:"window"`
}

type CORSConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Origins []string `mapstructure:"origins"`
	Methods []string `mapstructure:"methods"`
	Headers []string `mapstructure:"headers"`
	Expose  []string `mapstructure:"expose"`
}

type Server struct {
	*gin.Engine
	httpServer *http.Server
	config     *ServerConfig
}

func LoadServerConfig(config *config.Config) *ServerConfig {
	return &ServerConfig{
		Host:           config.Server.Host,
		Port:           config.Server.Port,
		ReadTimeout:    config.Server.ReadTimeout,
		WriteTimeout:   config.Server.WriteTimeout,
		IdleTimeout:    config.Server.IdleTimeout,
		MaxHeaderBytes: config.Server.MaxHeaderBytes,
		MaxConnections: config.Server.MaxConnections,
		MaxIdleConns:   config.Server.MaxIdleConns,
		RouterConfig: RouterConfig{
			Prefix: config.API.Prefix,
			RateLimit: RateLimitConfig{
				Enabled:     config.API.RateLimit.Enabled,
				MaxRequests: config.API.RateLimit.MaxRequests,
				Burst:       config.API.RateLimit.Burst,
				Window:      config.API.RateLimit.Window,
			},
			CORS: CORSConfig{
				Enabled: config.API.CORS.Enabled,
				Origins: strings.Split(config.API.CORS.AllowOrigin, " "),
				Methods: strings.Split(config.API.CORS.AllowMethods, " "),
				Headers: strings.Split(config.API.CORS.AllowHeaders, " "),
				Expose:  strings.Split(config.API.CORS.ExposeHeaders, " "),
			},
		},
	}
}

func NewServer(
	logger logger.Logger,
	config *config.Config,
) Server {
	appEnv := config.Server.Env
	if appEnv == "development" {
		gin.SetMode(gin.DebugMode)
	} else if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.TestMode)
	}
	serverConfig := LoadServerConfig(config)
	gin.DefaultWriter = logger.GetGinLogger()
	router := gin.New()
	router.Use(gin.Recovery())

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: logger.GetGinLogger(),
		Skip: func(c *gin.Context) bool {
			return c.Request.Method == http.MethodOptions
		},
		SkipPaths: []string{"/health", "/metrics"},
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s %s | %s %d | -> %s",
				param.TimeStamp.Format(time.RFC3339),
				param.ClientIP,
				param.Method,
				param.StatusCode,
				param.Path,
			)
		},
	}))

	// CORS middleware
	if serverConfig.RouterConfig.CORS.Enabled {
		setupCORS(router, serverConfig.RouterConfig.CORS)
	}

	// Rate limiting middleware
	if serverConfig.RouterConfig.RateLimit.Enabled {
		setupRateLimit(router, serverConfig.RouterConfig.RateLimit)
	}

	httpServer := &http.Server{
		Addr:           serverConfig.Host + ":" + fmt.Sprint(serverConfig.Port),
		Handler:        router,
		ReadTimeout:    config.Server.ReadTimeout,
		WriteTimeout:   config.Server.WriteTimeout,
		IdleTimeout:    config.Server.IdleTimeout,
		MaxHeaderBytes: serverConfig.MaxHeaderBytes,
	}
	// add prefix to all routes
	return Server{
		Engine:     router,
		httpServer: httpServer,
		config:     serverConfig,
	}
}

func setupCORS(router *gin.Engine, config CORSConfig) {
	if config.Enabled {
		router.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(config.Origins, ","))
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.Methods, ","))
			c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.Headers, ","))
			c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(config.Expose, ","))
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Next()
		})
	}
}

func setupRateLimit(router *gin.Engine, config RateLimitConfig) {
	if config.Enabled {
		router.Use(func(c *gin.Context) {
			// TODO: Implement rate limiting logic here
			c.Next()
		})
	}
}
