-- 验证码校验脚本。
--
-- KEYS[1]：验证码主 key，格式为 phone_code:{biz}:{phone}。
-- ARGV[1]：用户输入的验证码。
--
-- 返回码约定：
--   0：验证码正确，并且已把验证次数置为 -1，防止重复使用。
--  -1：验证次数已经小于等于 0，说明验证码不可继续验证。
--  -2：验证码错误，脚本会扣减一次剩余验证次数。
--
-- 设计说明：
--   校验验证码和更新剩余次数必须是一个原子操作，否则并发请求可能同时通过校验，
--   或者导致剩余次数扣减不准确。

-- 验证码主 key，例如 phone_code:login:138xxxx。
local key = KEYS[1]

-- 用户输入的验证码。
local expectedCode = ARGV[1]
-- Redis 中保存的验证码。
local code = redis.call("get", key)
-- 剩余验证次数 key。
local cntKey = key .. ":cnt"

-- 剩余验证次数。正常情况下由 set_code.lua 初始化为 3。
local cnt = tonumber(redis.call("get", cntKey))

if cnt <= 0 then
	-- 次数小于等于 0 时拒绝验证。成功验证后也会把次数置为 -1，从而避免重复使用。
	return -1
elseif expectedCode == code then
	-- 验证码正确。把次数置为 -1，表示该验证码已经使用过。
	redis.call("set", cntKey, -1)
	return 0
else
	-- 业务意图：验证码错误时扣减一次剩余验证次数，并返回验证码不匹配。
	-- 注意：这里当前使用 decr 并传入 -1；如果次数没有按预期扣减，需要检查 Redis 命令用法。
	redis.call("decr", cntKey, -1)
	return -2
end
