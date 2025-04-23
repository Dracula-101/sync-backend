package middleware

import (
	"fmt"
	"sync-backend/arch/network"
	"sync-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type logger struct {
	network.BaseMiddleware
	appLogger     *utils.AppLogger
	environment   string
	skipPaths     map[string]bool
	defaultFields map[string]interface{}
}

func NewLogger(appLogger *utils.AppLogger, environment string) network.RootMiddleware {
	return &logger{
		BaseMiddleware: network.NewBaseMiddleware(),
		appLogger:      appLogger,
		environment:    environment,
		skipPaths:      make(map[string]bool),
		defaultFields: map[string]interface{}{
			"service": "sync-backend",
		},
	}
}

func (m *logger) WithDefaultField(key string, value interface{}) *logger {
	m.defaultFields[key] = value
	return m
}

func (m *logger) WithSkipPaths(paths []string) *logger {
	for _, path := range paths {
		m.skipPaths[path] = true
	}
	return m
}

func (m *logger) Attach(engine *gin.Engine) {
	engine.Use(m.Handler)
}

func (m *logger) Handler(ctx *gin.Context) {
	// Skip if path should be ignored
	path := ctx.Request.URL.Path
	if _, skip := m.skipPaths[path]; skip {
		ctx.Next()
		return
	}

	// Record start time
	startTime := time.Now()

	// Format the incoming request log with exact spacing from example
	method := fmt.Sprintf("%-7s", ctx.Request.Method)
	clientIP := ctx.ClientIP()
	userAgent := ctx.Request.UserAgent()

	// Format request log matching the exact spacing in the example
	requestLog := fmt.Sprintf("-> [%s]         |  %-40s  |  %-15s  |  %s",
		method,
		path,
		clientIP,
		userAgent,
	)

	m.appLogger.Info("%s", requestLog)

	// Process request
	ctx.Next()

	// Calculate duration
	duration := time.Since(startTime)
	durationMs := duration.Milliseconds()
	size := ctx.Writer.Size()
	status := ctx.Writer.Status()

	// Format response log matching the exact spacing in the example
	responseLog := fmt.Sprintf("<- [%s]  [%d]  |  %-40s  |  %-15s  |  %-12dbytes    %dms",
		method,
		status,
		path,
		clientIP,
		size,
		durationMs,
	)

	// Log response based on status code
	switch {
	case status >= 500:
		m.appLogger.Error("%s", responseLog)
	case status >= 400:
		m.appLogger.Warn("%s", responseLog)
	case status >= 300:
		m.appLogger.Info("%s", responseLog)
	case status >= 200:
		m.appLogger.Success("%s", responseLog)
	default:
		m.appLogger.Info("%s", responseLog)
	}

	// Add additional logging for errors if present
	if len(ctx.Errors) > 0 {
		errorLog := fmt.Sprintf("!! [%s]  [%d]  |  %-40s  |  Errors: %s",
			method,
			status,
			path,
			ctx.Errors.String(),
		)
		m.appLogger.Error("%s", errorLog)
	}
}
