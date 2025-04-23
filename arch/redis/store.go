package redis

import (
	"context"
	"fmt"
	"sync-backend/utils"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host string
	Port uint16
	Pwd  string
	DB   int
}

type Store interface {
	GetInstance() *store
	Connect()
	Disconnect()
}

type store struct {
	*redis.Client
	logger  utils.AppLogger
	context context.Context
}

func NewStore(context context.Context, logger utils.AppLogger, config *Config) Store {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Pwd,
		DB:       config.DB,
	})
	return &store{
		context: context,
		logger:  logger,
		Client:  client,
	}
}

func (r *store) GetInstance() *store {
	return r
}

func (r *store) Connect() {
	r.logger.Debug("Connecting to redis...")
	r.logger.Debug("%s", fmt.Sprintf("Redis URI: %s:%d", r.Options().Addr, r.Options().DB))
	pong, err := r.Ping(r.context).Result()
	if err != nil {
		r.logger.Fatal("Could not connect to redis: %v", err)
	}
	r.logger.Success("Connected to redis: %s", pong)
}

func (r *store) Disconnect() {
	r.logger.Debug("Disconnecting from redis...")
	err := r.Close()
	if err != nil {
		r.logger.Fatal("Error disconnecting from redis: %v", err)
	}
	r.logger.Success("Disconnected from redis")
}
