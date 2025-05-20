package system

import (
	"fmt"
	"sync-backend/arch/network"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	network.BaseController
	service SystemService
}

func NewSystemController(service SystemService) *Controller {
	return &Controller{
		BaseController: network.NewBaseController("SystemController", nil),
		service:        service,
	}
}

func (c *Controller) Path() string {
	return "/api/v1/system"
}

func (c *Controller) MountRoutes(router *gin.RouterGroup) {
	router.GET("/status", c.GetStatus)
	router.GET("/health", c.GetHealth)
	router.GET("/routes", c.GetRoutes)
}

// GetStatus returns the overall system status
func (c *Controller) GetStatus(ctx *gin.Context) {
	status, err := c.service.GetSystemStatus()
	if err != nil {
		c.Send(ctx).InternalServerError(
			"Failed to get system status",
			fmt.Sprintf("Status check failed: %v - Service unavailable", err),
			network.UnknownErrorCode,
			err,
		)
		return
	}
	c.Send(ctx).SuccessDataResponse("System status retrieved successfully", status)
}

// GetHealth returns detailed health check information
func (c *Controller) GetHealth(ctx *gin.Context) {
	health, err := c.service.GetHealthStatus()
	if err != nil {
		c.Send(ctx).InternalServerError(
			"Failed to get system health",
			fmt.Sprintf("Health check failed: %v - Service unavailable", err),
			network.UnknownErrorCode,
			err,
		)
		return
	}
	c.Send(ctx).SuccessDataResponse("System health retrieved successfully", health)
}

// GetRoutes returns all registered API routes
func (c *Controller) GetRoutes(ctx *gin.Context) {
	routes, err := c.service.GetAPIRoutes()
	if err != nil {
		c.Send(ctx).InternalServerError(
			"Failed to get system routes",
			fmt.Sprintf("Route retrieval failed: %v - Service unavailable", err),
			network.UnknownErrorCode,
			err,
		)
		return
	}
	c.Send(ctx).SuccessDataResponse("System routes retrieved successfully", routes)
}
