package middleware

import (
	"sync-backend/arch/network"

	"github.com/gin-gonic/gin"
)

type methodNotAllowed struct {
	network.BaseMiddleware
}

func NewMethodNotAllowed() network.RootMiddleware {
	return &methodNotAllowed{
		BaseMiddleware: network.NewBaseMiddleware(),
	}
}

func (m *methodNotAllowed) Attach(engine *gin.Engine) {
	engine.NoMethod(m.Handler)
}

func (m *methodNotAllowed) Handler(ctx *gin.Context) {
	m.Send(ctx).MethodNotAllowedError("method not allowed", nil)
}