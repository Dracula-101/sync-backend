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
	case "debug":
	case "development":
		mode = gin.DebugMode
	case "staging":
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
	eng.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return LoggerFormatter(appLogger, gin.DebugMode == "debug")(param)
		},
	}))
	eng.Use(gin.Recovery())
	eng.Use(gin.ErrorLogger())

	eng.HandleMethodNotAllowed = true
	eng.NoMethod(NotAllowed())
	eng.NoRoute(NotFound())
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
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		return 0, nil
	}
	switch w.level {
	case "info":
		w.appLogger.Info("[GIN-%s]: %s", w.level, s)
	case "debug":
		w.appLogger.Debug("[GIN-%s]: %s", w.level, s)
	case "warn":
		w.appLogger.Warn("[GIN-%s]: %s", w.level, s)
	case "error":
		w.appLogger.Error("[GIN-%s]: %s", w.level, s)
	default:
		w.appLogger.Info("[GIN-%s]: %s", w.level, s)
	}
	return len(p), nil
}
