package network

type baseController struct {
	ResponseSender
	basePath string
}

func NewBaseController(basePath string) BaseController {
	return &baseController{
		ResponseSender: NewResponseSender(),
		basePath:       basePath,
	}
}

func (c *baseController) Path() string {
	return c.basePath
}

// func (c *baseController) Authentication() gin.HandlerFunc {
// 	return c.authProvider.Middleware()
// }

// func (c *baseController) Authorization(role string) gin.HandlerFunc {
// 	return c.authorizeProvider.Middleware(role)
// }
