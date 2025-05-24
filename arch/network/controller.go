package network

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type baseController struct {
	ResponseSender
	basePath     string
	authProvider AuthenticationProvider
}

func NewBaseController(basePath string, authProvider AuthenticationProvider) BaseController {
	return &baseController{
		ResponseSender: NewResponseSender(),
		basePath:       basePath,
		authProvider:   authProvider,
	}
}

func (c *baseController) Path(basePath string) string {
	return fmt.Sprintf("%s%s", basePath, c.basePath)
}

func (c *baseController) Authentication() gin.HandlerFunc {
	return c.authProvider.Middleware()
}
