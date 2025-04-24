package application

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"sync-backend/arch/config"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/arch/redis"
	"sync-backend/utils"

	"github.com/gin-gonic/gin"
)

type Shutdown = func()

func Server() {
	env := config.NewEnv(".env")
	config := config.LoadConfig("./configs")
	router, _, shutdown := create(&env, &config)
	defer shutdown()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	go func() {
		router.Start(env.Host, uint16(env.Port))
	}()
	<-stop
}

func create(env *config.Env, config *config.Config) (network.Router, Module, Shutdown) {
	context := context.Background()

	serverLogger := utils.DefaultAppLogger(env.Env, env.LogLevel, "Server")
	dbLogger := utils.DefaultAppLogger(env.Env, env.LogLevel, "Database")
	redisLogger := utils.DefaultAppLogger(env.Env, env.LogLevel, "Redis")

	dbConfig := mongo.DbConfig{
		User:        env.DBUser,
		Pwd:         env.DBPassword,
		Host:        env.DBHost,
		Name:        env.DBName,
		MinPoolSize: uint16(config.DB.MinPoolSize),
		MaxPoolSize: uint16(config.DB.MaxPoolSize),
		Timeout:     config.DB.TimeoutConfig.ConnectTimeout,
	}

	db := mongo.NewDatabase(context, dbLogger, dbConfig)
	db.Connect()

	if env.Env != gin.TestMode {
		EnsureDbIndexes(db)
	}

	redisConfig := redis.Config{
		Host: env.RedisHost,
		Port: uint16(env.RedisPort),
		Pwd:  env.RedisPassword,
		DB:   env.RedisDB,
	}

	store := redis.NewStore(context, redisLogger, &redisConfig)
	store.Connect()

	module := NewAppModule(context, env, config, db, store)
	router := network.NewRouter(env.Env, serverLogger)
	router.RegisterValidationParsers(network.CustomTagNameFunc())
	router.LoadRootMiddlewares(module.RootMiddlewares())
	router.LoadControllers(module.Controllers())

	shutdown := func() {
		db.Disconnect()
		store.Disconnect()
	}

	return router, module, shutdown
}
