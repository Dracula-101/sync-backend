package middleware

import (
	"fmt"
	"runtime/debug"
	"sync-backend/arch/network"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type errorCatcher struct {
	network.BaseMiddleware
	logger *utils.AppLogger
}

func NewErrorCatcher(logger *utils.AppLogger) network.RootMiddleware {
	return &errorCatcher{
		BaseMiddleware: network.NewBaseMiddleware(),
		logger:         logger,
	}
}

func (m *errorCatcher) Attach(engine *gin.Engine) {
	engine.Use(m.Handler)
}

func (m *errorCatcher) Handler(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			// Get stack trace
			stackTrace := debug.Stack()

			// Log the error with stack trace
			errorMsg := fmt.Sprintf("PANIC RECOVERED [%s %s]: %v\n%s",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				r,
				string(stackTrace))

			// Log to standard logger
			if m.logger != nil {
				m.logger.Error("%s", errorMsg)
			}

			// Return appropriate response to client
			if err, ok := r.(error); ok {
				m.Send(ctx).InternalServerError(err.Error(), err)
			} else {
				m.Send(ctx).InternalServerError("something went wrong", err)
			}

			ctx.Abort()
		}
	}()
	ctx.Next()
}
