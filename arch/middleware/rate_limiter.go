package middleware

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync-backend/arch/config"
	"sync-backend/arch/network"
	"sync-backend/arch/redis"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	network.BaseMiddleware
	redis  redis.Store
	config config.Config
}

func NewRateLimiter(redis redis.Store, config config.Config) network.RootMiddleware {
	return &rateLimiter{
		BaseMiddleware: network.NewBaseMiddleware(),
		redis:          redis,
		config:         config,
	}
}

func (m *rateLimiter) Attach(engine *gin.Engine) {
	engine.Use(m.Handler)
}

func (m *rateLimiter) Handler(ctx *gin.Context) {
	ip := ctx.ClientIP()

	// Add timestamp to create a time-bound window key
	windowStart := time.Now().Unix() / int64(m.config.Auth.RateLimit.General.Duration.Seconds()) * int64(m.config.Auth.RateLimit.General.Duration.Seconds())
	key := fmt.Sprintf("ratelimit:ip:%s:%d", ip, windowStart)

	limit := m.config.Auth.RateLimit.General.Requests
	windowSeconds := m.config.Auth.RateLimit.General.Duration.Seconds()
	if limit <= 0 || windowSeconds <= 0 {
		ctx.Next()
		return
	}

	// Using a pipeline to make operations atomic
	pipe := m.redis.GetInstance().Pipeline()
	incr := pipe.Incr(context.Background(), key)
	pipe.Expire(context.Background(), key, time.Duration(windowSeconds)*time.Second)
	_, err := pipe.Exec(context.Background())

	if err != nil {
		ctx.Next()
		return
	}

	val := incr.Val()

	// Calculate time until window resets
	timeUntilReset := windowStart + int64(windowSeconds) - time.Now().Unix()
	if timeUntilReset < 0 {
		timeUntilReset = int64(windowSeconds)
	}

	ctx.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	ctx.Header("X-RateLimit-Remaining", strconv.Itoa(int(math.Max(0, float64(limit)-float64(val)))))
	ctx.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Unix()+timeUntilReset, 10))

	if val > int64(limit) {
		m.Send(ctx).TooManyRequestsError(
			"Rate limit exceeded",
			fmt.Sprintf("Rate limit exceeded. Try again in %d seconds", timeUntilReset),
			fmt.Errorf("Rate limit exceeded for IP %s: %d requests in %.0f seconds", ip, val, windowSeconds),
		)
		return
	}

	ctx.Next()
}
