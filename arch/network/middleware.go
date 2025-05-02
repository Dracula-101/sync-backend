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
		c.JSON(405, NewEnvelopeWithErrors(
			false,
			405,
			"Method Not Allowed",
			[]ErrorDetail{
				{
					Code:    "METHOD_NOT_ALLOWED",
					Message: fmt.Sprintf("%s Method Not Allowed", c.Request.Method),
					Detail:  fmt.Sprintf("The method %s is not allowed for the requested URL %s", c.Request.Method, c.Request.URL),
				},
			},
		))
		c.Abort()
	}
}

func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(404, NewEnvelopeWithErrors(
			false,
			404,
			"Method Not Found",
			[]ErrorDetail{
				{
					Code:    "NOT_FOUND",
					Message: fmt.Sprintf("%s Not Found", c.Request.Method),
					Detail:  fmt.Sprintf("The requested URL %s was not found on this server", c.Request.URL),
				},
			},
		))
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
