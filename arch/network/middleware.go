package network

import "github.com/gin-gonic/gin"

type baseMiddleware struct {
	ResponseSender
}

func NewBaseMiddleware() BaseMiddleware {
	return &baseMiddleware{
		ResponseSender: NewResponseSender(),
	}
}

func (m *baseMiddleware) Debug() bool {
	return gin.Mode() == gin.DebugMode
}

func NotAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(405, gin.H{
			"message": "Method Not Allowed",
			"status":  405,
		})
		c.Abort()
	}
}

func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(404, gin.H{
			"message": "Url Not Found",
			"status":  404,
		})
		c.Abort()
	}
}
