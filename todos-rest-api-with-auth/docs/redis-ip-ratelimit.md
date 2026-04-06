# Redis IP 限流实现文档

## 概述

本项目采用 **Redis 滑动窗口算法** 实现基于 IP 的请求限流，保护 API 服务免受过度请求和恶意攻击。

## 算法原理

### 滑动窗口 vs 固定窗口

**固定窗口算法** 在每个固定时间窗口（如 1 分钟）内允许固定数量的请求。缺点是窗口边界可能出现突发流量。

**滑动窗口算法** 使用时间戳来动态计算窗口内的请求数，比固定窗口更精确。

### 数据结构

Redis 有序集合（Sorted Set）用于存储请求记录：

- **Key**: 限流标识（如 `ip-limiter:192.168.1.1`）
- **Member**: 请求唯一 ID（UUID）
- **Score**: 请求时间戳（毫秒）

### 限流流程

1. **清理窗口外数据**: 删除窗口起始时间之前的所有请求记录
2. **统计当前窗口**: 统计窗口内的请求数量
3. **判断限流**: 如果数量 >= 阈值，执行限流；否则记录本次请求并放行

## 核心实现

### Lua 脚本（原子性保证）

```lua
--[[
    滑动窗口限流 Lua 脚本

    算法原理:
    使用 Redis 有序集合 (Sorted Set) 存储每个请求的时间戳。
    - member: 请求的唯一标识 (UUID)
    - score: 请求发生的时间戳 (毫秒)

    限流判断:
    统计窗口内 (当前时间 - 窗口大小) 到 当前时间 范围内的请求数量。
    如果数量 >= 阈值，则限流。

    参数说明:
    KEYS[1]: 限流 key (如用户ID、IP等)
    ARGV[1]: 窗口大小 (毫秒)
    ARGV[2]: 限流阈值 (最大请求数)
    ARGV[3]: 当前时间戳 (毫秒)
    ARGV[4]: 请求唯一ID (UUID)

    返回值:
    "true"  - 请求被限流，应拒绝
    "false" - 请求允许通过

    优势:
    1. 滑动窗口比固定窗口更精确，避免窗口边界突发流量问题
    2. Lua 脚本保证原子性，避免并发问题
    3. 使用 PEXPIRE 自动清理过期数据，防止内存泄漏
]]

local key = KEYS[1]
local window = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local uid = ARGV[4]

-- 计算窗口起始时间
local min = now - window

-- 步骤1: 移除窗口外的过期请求
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)

-- 步骤2: 统计当前窗口内的请求数量
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

-- 步骤3: 判断是否需要限流
if cnt >= threshold then
    return "true"
else
    redis.call('ZADD', key, now, uid)
    redis.call('PEXPIRE', key, window)
    return "false"
end
```

### Go 限流器实现

```go
// RedisSlidingWindowLimiter 是基于 Redis 的滑动窗口限流器。
type RedisSlidingWindowLimiter struct {
    Cmd      redis.Cmdable    // Redis 客户端
    Interval time.Duration    // 窗口时间范围
    Rate     int              // 允许的最大请求数
}

// Limit 判断给定 key 的请求是否应该被限流。
func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
    uid, err := uuid.NewUUID()
    if err != nil {
        return false, fmt.Errorf("generate uuid failed: %v", err)
    }

    // 执行 Lua 脚本
    return r.Cmd.Eval(ctx, luaSlideWindow, []string{key},
        r.Interval.Milliseconds(), r.Rate, time.Now().UnixMilli(), uid.String()).Bool()
}
```

## 中间件集成

### 快速使用

```go
import (
    "time"
    "github.com/redis/go-redis/v9"
    "todo_api/internal/ratelimit"
    ratelimitmiddleware "todo_api/internal/middleware/ratelimit"
)

// 创建 Redis 客户端
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 创建滑动窗口限流器（1分钟内最多5次请求）
limiter := ratelimit.NewRedisSlidingWindowLimiter(redisClient, time.Minute, 5)

// 使用构建器创建中间件
rateMiddleware := ratelimitmiddleware.NewBuilder(limiter).Build()

// 注册到路由
router.POST("/auth/login", rateMiddleware, handlers.LoginHandler(pool, cfg))
```

### 自定义配置

```go
// 自定义限流 Key 生成函数
rateMiddleware := ratelimitmiddleware.NewBuilder(limiter).
    SetKeyGenFunc(func(ctx *gin.Context) string {
        // 按用户ID限流
        userID := ctx.GetHeader("X-User-ID")
        return "user-limit:" + userID
    }).
    SetLogFunc(func(msg any, args ...any) {
        // 自定义日志
        log.Printf("[RateLimit] %v %v", msg, args)
    }).
    Build()
```

## 关键设计

### 1. 原子性保证

Lua 脚本在 Redis 中原子执行，避免并发请求导致的竞态条件。

### 2. 内存管理

使用 `PEXPIRE` 自动设置过期时间，确保窗口滑动后旧数据被清理，防止内存泄漏。

### 3. 高精度

UUID 解决同一毫秒内多个请求的计数问题，确保限流计数准确。

### 4. 可替换性

`Limiter` 接口设计使不同限流算法可以轻松替换：

```go
type Limiter interface {
    Limit(ctx context.Context, key string) (bool, error)
}
```

## 配置建议

| 场景 | 窗口时间 | 请求限额 |
|------|---------|---------|
| 登录接口 | 1 分钟 | 5 次 |
| 注册接口 | 1 分钟 | 3 次 |
| 普通 API | 1 分钟 | 60 次 |
| 搜索接口 | 1 分钟 | 30 次 |

## 限流响应

当请求被限流时，返回 HTTP 429：

```json
{
    "message": "Too Many Requests"
}
```

## 监控指标

建议监控以下指标：

- `ratelimit_blocked_total`: 被限流的请求总数
- `ratelimit_error_total`: 限流错误总数
- `ratelimit_latency_seconds`: 限流延迟分布

## 底层算法详解

### 滑动窗口算法数学原理

滑动窗口算法的核心是将时间轴划分为连续的时间窗口，并通过统计窗口内的请求数来判断是否限流。

**传统固定窗口算法的问题：**

```
时间轴: |-----T1-----|-----T2-----|-----T3-----|
请求:     ↑  ↑  ↑       ↑  ↑  ↑  ↑  ↑
          3个请求      4个请求

问题: 在 T1 末尾和 T2 开始时刻可能产生 7 个请求的突发
```

**滑动窗口算法的改进：**

```
时间轴: |-----------------滑动窗口-----------------|
窗口大小: W 毫秒
当前时刻: now
窗口范围: [now - W, now]

窗口随时间连续滑动，统计的是任意时刻的最近 W 时间内的请求数
```

### Redis 有序集合操作解析

本实现使用 Redis ZSET（有序集合）存储请求时间戳，每条记录：

```
Key:    "ip-limiter:192.168.1.1"
Member: "uuid-xxx-xxx"        (唯一请求ID)
Score:  1743849600000         (时间戳毫秒)
```

**ZREMRANGEBYSCORE**

删除窗口外的过期记录：

```lua
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
-- 删除所有 score < min 的元素
-- min = now - window
```

**ZCOUNT**

统计窗口内的请求数：

```lua
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')
-- 统计 score 在 (-∞, +∞) 范围内的元素数量
```

**ZADD + PEXPIRE**

记录新请求并设置过期：

```lua
redis.call('ZADD', key, now, uid)        -- 添加新请求
redis.call('PEXPIRE', key, window)       -- 设置过期时间
```

### 算法复杂度

| 操作 | 时间复杂度 | 说明 |
|------|-----------|------|
| ZREMRANGEBYSCORE | O(log(N)+M) | N=总记录数，M=删除数 |
| ZCOUNT | O(log(N)) | N=窗口内记录数 |
| ZADD | O(log(N)) | N=总记录数 |

其中 N 受窗口大小和请求速率影响，实际应用中 N 较小。

### 并发安全性

Redis Lua 脚本保证原子性：

```lua
-- 整个限流逻辑在一个 Lua 脚本中执行
-- Redis 执行 Lua 脚本时不会执行其他命令
-- 避免了 GET + SET 分离导致的竞态条件
```

### 与其他算法对比

| 算法 | 精确度 | 内存占用 | 实现复杂度 | 适用场景 |
|------|--------|----------|------------|----------|
| 固定窗口 | 低 | 低 | 简单 | 粗粒度限流 |
| 滑动窗口 | 中 | 中 | 中等 | 大多数场景 |
| 令牌桶 | 高 | 低 | 复杂 | 平滑限流 |
| 漏桶 | 高 | 低 | 复杂 | 严格限流 |

滑动窗口在精确度和实现复杂度之间取得平衡，是大多数 API 限流的理想选择。

### 边界情况处理

1. **Redis 宕机**: 返回 error，取决于业务策略（放行或限流）
2. **时钟偏移**: 使用 Redis 服务器时间，避免客户端时钟不同步
3. **同一毫秒多请求**: UUID 作为 member 区分，同一时间戳可有多条记录
4. **窗口大小为0**: 理论不支持，代码中 interval 必须 > 0

## 常见问题

**Q: 限流功能不生效？**

检查：
1. Redis 服务是否运行
2. Redis 连接地址是否正确
3. 限流 key 是否被正确生成

**Q: 如何调整限流阈值？**

修改 `NewRedisSlidingWindowLimiter` 的 `rate` 参数：

```go
// 1分钟内允许 100 次请求
limiter := ratelimit.NewRedisSlidingWindowLimiter(redisClient, time.Minute, 100)
```

**Q: 如何实现分布式限流？**

确保所有服务实例连接同一个 Redis 集群即可，限流状态存储在 Redis 中全局共享。
