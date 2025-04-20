package network

import (
	"fmt"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type router struct {
	engine *gin.Engine
}

func NewRouter(env string, appLogger utils.AppLogger) Router {
	var mode string
	switch env {
	case "development":
		mode = gin.DebugMode
	case "production":
		mode = gin.ReleaseMode
	case "test":
		mode = gin.TestMode
	}
	gin.SetMode(mode)
	if gin.DebugMode == mode {
		gin.DefaultWriter = &logWriter{appLogger: appLogger, level: "info"}
		gin.DefaultErrorWriter = &logWriter{appLogger: appLogger, level: "error"}
	}
	eng := gin.New()
	eng.Use(gin.Logger())
	eng.Use(gin.Recovery())
	eng.Use(gin.ErrorLogger())

	eng.HandleMethodNotAllowed = true
	eng.NoMethod(NotFound())
	eng.NoRoute(NotAllowed())
	r := router{
		engine: eng,
	}
	return &r
}

func (r *router) GetEngine() *gin.Engine {
	return r.engine
}

func (r *router) LoadRootMiddlewares(middlewares []RootMiddleware) {
	for _, m := range middlewares {
		m.Attach(r.engine)
	}
}

func (r *router) LoadControllers(controllers []Controller) {
	for _, c := range controllers {
		g := r.engine.Group(c.Path())
		c.MountRoutes(g)
	}
}

func (r *router) Start(ip string, port uint16) {
	address := fmt.Sprintf("%s:%d", ip, port)
	r.engine.Run(address)
}

func (r *router) RegisterValidationParsers(tagNameFunc validator.TagNameFunc) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(tagNameFunc)
	}
}

type logWriter struct {
	appLogger utils.AppLogger
	level     string
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	s := string(p)

	// Remove trailing newlines if present
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}

	// Skip empty strings
	if len(s) == 0 {
		return len(p), nil
	}

	// Log at appropriate level
	switch w.level {
	case "info":
		w.appLogger.Info("GIN: %s", s)
	case "debug":
		w.appLogger.Debug("GIN: %s", s)
	case "warn":
		w.appLogger.Warn("GIN: %s", s)
	case "error":
		w.appLogger.Error("GIN: %s", s)
	default:
		w.appLogger.Info("GIN: %s", s)
	}

	return len(p), nil
}
