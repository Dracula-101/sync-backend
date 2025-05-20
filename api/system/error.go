package system

import "sync-backend/arch/network"

const (
	ERR_DATABASE = "ERR_DATABASE"
	ERR_REDIS    = "ERR_REDIS"
	ERR_SYSTEM   = "ERR_SYSTEM"
)

var (
	ErrDatabaseUnavailable = network.NewInternalServerError("Database is unavailable", ERR_DATABASE, nil)
	ErrRedisUnavailable    = network.NewInternalServerError("Redis is unavailable", ERR_REDIS, nil)
	ErrSystemDegraded      = network.NewInternalServerError("System is in degraded state", ERR_SYSTEM, nil)
) 