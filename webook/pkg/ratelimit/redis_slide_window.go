package ratelimit

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaSlideWindow string

type RedisSlidingWindowLimiter struct {
	cmd redis.Cmdable

	interval time.Duration

	rate int
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	uid, err := uuid.NewUUID()
	if err != nil {
		return false, fmt.Errorf("generate uuid failed: %v", err)
	}
	return r.cmd.Eval(ctx, luaSlideWindow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli(), uid.String()).Bool()
}
