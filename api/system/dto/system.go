package dto

import "time"

// SystemStatusResponse represents the overall system status response
type SystemStatusResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

// APIRouteResponse represents a registered API route
type APIRouteResponse struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}

// SystemLogResponse represents a system log entry
type SystemLogResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}
