package middleware

import (
	"sync-backend/arch/network"

	"github.com/gin-gonic/gin"
)

type notFound struct {
	network.BaseMiddleware
}

func NewNotFound() network.RootMiddleware {
	return &notFound{
		BaseMiddleware: network.NewBaseMiddleware(),
	}
}

func (m *notFound) Attach(engine *gin.Engine) {
	engine.NoRoute(m.Handler)
}

func (m *notFound) Handler(ctx *gin.Context) {
	m.Send(ctx).NotFoundError("url not found", nil)
}
