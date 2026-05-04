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

-- 限流 key，标识被限流的资源或用户
local key = KEYS[1]

-- 窗口大小 (毫秒)，时间窗口的范围
local window = tonumber(ARGV[1])

-- 限流阈值，时间窗口内允许的最大请求数
local threshold = tonumber(ARGV[2])

-- 当前时间戳 (毫秒)
local now = tonumber(ARGV[3])

-- 请求唯一ID，使用 UUID 解决同一毫秒内多个请求的计数问题
local uid = ARGV[4]

-- 计算窗口起始时间
local min = now - window

-- 步骤1: 移除窗口外的过期请求
-- ZREMRANGEBYSCORE 删除 score 小于窗口起始时间的所有元素
-- 这里用 '-inf' 表示负无穷，即删除所有 min 之前的请求
redis.call("ZREMRANGEBYSCORE", key, "-inf", min)

-- 步骤2: 统计当前窗口内的请求数量
local cnt = redis.call("ZCOUNT", key, "-inf", "+inf")

-- 步骤3: 判断是否需要限流
if cnt >= threshold then
	-- 超过阈值，执行限流
	return "true"
else
	-- 未超过阈值，记录本次请求并放行
	-- ZADD 添加元素，score 为时间戳，member 为唯一ID
	redis.call("ZADD", key, now, uid)
	-- PEXPIRE 设置过期时间，自动清理数据防止内存泄漏
	-- 过期时间设为窗口大小，确保窗口滑动后数据被清理
	redis.call("PEXPIRE", key, window)
	return "false"
end
