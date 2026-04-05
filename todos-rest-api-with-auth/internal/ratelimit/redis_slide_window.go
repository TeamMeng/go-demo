// Package ratelimit 提供基于 Redis 滑动窗口算法的限流功能。
// 使用在 Redis 中执行的 Lua 脚本保证原子操作，在高并发场景下确保准确性。
package ratelimit

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// luaSlideWindow 是编译时嵌入的 Lua 脚本，用于 Redis 执行。
// 脚本实现滑动窗口限流算法，保证操作的原子性。
//
//go:embed slide_window.lua
var luaSlideWindow string

// RedisSlidingWindowLimiter 是基于 Redis 的滑动窗口限流器。
// 使用有序集合追踪时间窗口内的请求时间戳，比固定窗口算法更精确。
type RedisSlidingWindowLimiter struct {
	Cmd redis.Cmdable

	// Interval 是滑动窗口的时间范围。
	// 窗口内的请求计入限流计数。
	Interval time.Duration

	// Rate 是时间窗口内允许的最大请求数。
	Rate int
}

// Limit 判断给定 key 的请求是否应该被限流。
// 返回 true 表示请求被限流（超过阈值），false 表示允许通过。
// key 用于标识被限流的资源或用户。
func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	uid, err := uuid.NewUUID()
	if err != nil {
		return false, fmt.Errorf("generate uuid failed: %v", err)
	}

	// 在 Redis 中执行 Lua 脚本，实现原子性滑动窗口限流。
	// 参数：key, 窗口时间(毫秒), 限流阈值, 当前时间(毫秒), 唯一请求ID。
	return r.Cmd.Eval(ctx, luaSlideWindow, []string{key}, r.Interval.Milliseconds(), r.Rate, time.Now().UnixMilli(), uid.String()).Bool()
}
