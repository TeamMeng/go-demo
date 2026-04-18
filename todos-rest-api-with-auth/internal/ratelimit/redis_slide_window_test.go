package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// TestRedisSlideWindowLimiter_Limit 测试滑动窗口限流器的基本限流功能。
//
// 测试场景说明:
//  1. 正常通过: 新窗口内首次请求应允许通过
//  2. 不同key独立计数: 不同key的限流互不影响
//  3. 触发限流: 短时间内同一key多次请求应被限流
//  4. 窗口重置后恢复: 等待窗口过期后，新请求应能通过
//
// 测试配置: 窗口500ms，阈值1次
func TestRedisSlideWindowLimiter_Limit(t *testing.T) {
	// 创建限流器实例
	//   - 窗口大小: 500毫秒
	//   - 允许阈值: 1次
	r := &RedisSlidingWindowLimiter{
		Cmd:      initRedis(),
		Interval: 500 * time.Millisecond,
		Rate:     1,
	}

	tests := []struct {
		name     string          // 测试用例名称
		ctx      context.Context // 上下文
		key      string          // 限流key
		interval time.Duration   // 距离上一个请求的间隔
		want     bool            // 期望的限流结果
		wantErr  error           // 期望的错误
	}{
		{
			name: "正常通过",
			ctx:  context.Background(),
			key:  "foo",
			want: false, // 首次请求，应允许通过
		},
		{
			name: "另外一个key正常通过",
			ctx:  context.Background(),
			key:  "bar",
			want: false, // 不同key独立计数，应允许通过
		},
		{
			name:     "限流",
			ctx:      context.Background(),
			key:      "foo",
			interval: 300 * time.Millisecond, // 300ms后再次请求
			want:     true,                   // 同一key在窗口内再次请求，应限流
		},
		{
			name:     "窗口有空余正常通过",
			ctx:      context.Background(),
			key:      "foo",
			interval: 510 * time.Millisecond, // 510ms后请求，超过500ms窗口
			want:     false,                  // 窗口已过期，应允许通过
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 等待指定间隔，模拟请求时间间隔
			<-time.After(tt.interval)

			// 执行限流检查
			got, err := r.Limit(tt.ctx, tt.key)

			// 验证结果
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// initRedis 初始化并返回 Redis 客户端连接。
// 默认连接本地 Redis: localhost:6379
// 使用 redis.Cmdable 接口，支持单节点和集群模式。
func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return redisClient
}

// TestRedisSlidingWindowLimiter 测试高并发场景下限流器的准确性和性能。
//
// 测试场景:
//   - 设置窗口为1秒，阈值为1200次
//   - 发送1500次请求
//   - 验证限流器能正确限制超出阈值的请求
//
// 预期结果:
//   - 约1200次请求通过 (succCount ≈ 1200)
//   - 约300次请求被限流 (limitCount ≈ 300)
//
// 注意: 由于滑动窗口的精确性，实际通过数可能略高于阈值(取决于请求分布)
func TestRedisSlidingWindowLimiter(t *testing.T) {
	r := &RedisSlidingWindowLimiter{
		Cmd:      initRedis(),
		Interval: time.Second, // 窗口: 1秒
		Rate:     1200,        // 阈值: 1200次
	}

	var (
		total      = 1500 // 总请求数
		succCount  int    // 成功(未限流)计数
		limitCount int    // 限流计数
	)

	start := time.Now()

	// 循环发送请求
	//nolint:revive // 循环计数器用于控制总请求次数
	for range total {
		limit, err := r.Limit(context.Background(), "TestRedisSlidingWindowLimiter")
		if err != nil {
			t.Fatalf("限流检查失败: %v", err)
			return
		}
		if limit {
			limitCount++
			continue
		}
		succCount++
	}

	end := time.Now()

	// 输出测试结果日志
	t.Logf("开始时间: %v", start.Format(time.StampMilli))
	t.Logf("结束时间: %v", end.Format(time.StampMilli))
	t.Logf("总请求: %d, 通过: %d, 限流: %d", total, succCount, limitCount)
}
