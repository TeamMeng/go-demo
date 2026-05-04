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
	Cmd      redis.Cmdable
	Interval time.Duration
	Rate     int
}

func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	uid, err := uuid.NewUUID()
	if err != nil {
		return false, fmt.Errorf("generate uuid failed: %v", err)
	}
	return r.Cmd.Eval(ctx, luaSlideWindow, []string{key},
		r.Interval.Milliseconds(), r.Rate, time.Now().UnixMilli(), uid.String()).Bool()
}
