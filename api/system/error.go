package system

import (
	"sync-backend/arch/network"
)

const (
	ERR_DATABASE = "ERR_DATABASE"
	ERR_REDIS    = "ERR_REDIS"
	ERR_SYSTEM   = "ERR_SYSTEM"
)

var (
	ErrDatabaseUnavailable = network.NewInternalServerError(
		"Database is unavailable",
		"Database is unavailable. Please check the database connection and configuration.",
		ERR_DATABASE,
		nil,
	)
	ErrRedisUnavailable = network.NewInternalServerError(
		"Redis is unavailable",
		"Redis is unavailable. Please check the Redis connection and configuration.",
		ERR_REDIS,
		nil,
	)
	ErrSystemDegraded = network.NewInternalServerError(
		"System is degraded",
		"System is degraded. Some features may not work as expected.",
		ERR_SYSTEM,
		nil,
	)
)
