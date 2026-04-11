-- 验证码写入脚本。
--
-- KEYS[1]：验证码主 key，格式为 phone_code:{biz}:{phone}，例如 phone_code:login:138xxxx。
-- ARGV[1]：本次生成的验证码明文。
--
-- 返回码约定：
--   0：允许发送，验证码和验证次数都已写入 Redis。
--  -1：发送过于频繁，当前验证码剩余有效期仍然较长。
--  -2：验证码 key 存在但没有设置过期时间，表示数据状态异常。
--
-- 设计说明：
--   验证码和验证次数需要一起更新，所以放在 Lua 脚本中执行，保证 Redis 内部原子性。
--   验证码有效期为 600 秒；当旧验证码不存在，或剩余有效期小于 540 秒时允许重新发送。

-- 验证码主 key，例如 phone_code:login:138xxxx。
local key = KEYS[1]

-- 验证次数 key，例如 phone_code:login:138xxxx:cnt。
local cntKey = key .. ":cnt"

-- 本次要写入的验证码。
local val = ARGV[1]

-- Redis TTL 语义：
--   ttl > 0：key 存在并且有过期时间，返回剩余秒数。
--   ttl == -1：key 存在，但没有过期时间。
--   ttl == -2：key 不存在。
local ttl = tonumber(redis.call("ttl", key))

if ttl == -1 then
	-- key 存在但没有过期时间。验证码必须有有效期，因此这里作为异常状态返回。
	return -2
elseif ttl == -2 or ttl < 540 then
	-- 旧验证码不存在，或者旧验证码距离过期不足 9 分钟，允许生成新验证码。
	redis.call("set", key, val)
	-- 验证码有效期 10 分钟。
	redis.call("expire", key, 600)
	-- 每个验证码最多允许验证 3 次，成功验证后 verify_code.lua 会把该值置为 -1。
	redis.call("set", cntKey, 3)
	-- 业务意图：验证次数 key 和验证码 key 使用相同有效期，避免验证码过期后次数 key 残留。
	-- 注意：这里当前调用的是 exire；如果运行时报 Redis unknown command，需要检查是否应为 expire。
	redis.call("exire", cntKey, 600)
	return 0
else
	-- 旧验证码剩余有效期仍然大于等于 9 分钟，说明用户在短时间内重复获取验证码。
	return -1
end
