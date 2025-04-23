package network

import (
	"fmt"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

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

func LoggerFormatter(appLogger utils.AppLogger, debug bool) gin.LogFormatter {
	return func(param gin.LogFormatterParams) string {
		message := fmt.Sprintf("%s  | %6s | %3v | %15s | %-7s | %d-bytes | %#v %s",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.BodySize,
			param.Path,
			param.ErrorMessage,
		)
		switch {
		case param.StatusCode >= 500:
			appLogger.Error("%s", message)
		case param.StatusCode >= 300:
			appLogger.Warn("%s", message)
		case param.StatusCode >= 200:
			appLogger.Success("%s", message)
		default:
			appLogger.Info("%s", message)
		}
		return ""
	}
}
