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
	logger utils.AppLogger
}

func NewErrorCatcher() network.RootMiddleware {
	return &errorCatcher{
		BaseMiddleware: network.NewBaseMiddleware(),
		logger:         utils.NewServiceLogger("ErrorCatcher"),
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

			m.logger.Error("%s", errorMsg)

			// Return appropriate response to client
			if err, ok := r.(error); ok {
				m.Send(ctx).InternalServerError(
					err.Error(),
					"Server encountered an expected error and cannot process the request [Context - MAJOR ERROR]",
					network.UnknownErrorCode,
					err,
				)
			} else {
				m.Send(ctx).InternalServerError(
					"Something went wrong",
					"Server encountered an expected error and cannot process the request",
					network.UnknownErrorCode,
					nil,
				)
			}

			ctx.Abort()
		}
	}()
	ctx.Next()
}
