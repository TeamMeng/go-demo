--[[
	验证码校验脚本。

	Lua 脚本的核心价值：原子化地完成「读取 → 比较 → 扣减次数」，
	避免并发请求下出现「同时通过校验」或「次数扣减不一致」的问题。

	KEYS[1]: 验证码主 key，格式为 phone_code:{biz}:{phone}
	ARGV[1]:  用户输入的验证码明文

	返回码约定：
	  0: 验证码正确，并且把验证次数置为 -1，防止重复使用
	 -1: 剩余验证次数已耗尽（<= 0），拒绝验证
	 -2: 验证码不存在或已过期，或验证码错误（两者合并返回，调用方无需区分）
--]]

local key = KEYS[1]
local inputCode = ARGV[1]

-- 读取验证码本体
local storedCode = redis.call("get", key)
-- 读取剩余验证次数
local cntKey = key .. ":cnt"
local cntVal = redis.call("get", cntKey)

--[[
	防御性检查：任一 key 不存在或异常，都按「验证码无效」处理。

	- storedCode == false : 验证码已过期（TTL 归零后 Redis 自动删除）
	- cntVal == false     : 次数 key 过期，或从未创建过（理论上应由 set_code.lua 保障）
	- tonumber 失败        : 数据被外部错误写入非数字值

	为什么不区分返回码？
	  站在安全角度，验证码「不存在」和「已过期」对于攻击者没有本质区别，
	  都应该引导用户重新获取验证码。合并返回可以简化上层判断逻辑。
]]
if storedCode == false or cntVal == false then
	return -2
end

local cnt = tonumber(cntVal)
if cnt == nil then
	-- 非数字值按无效处理，避免 Lua 后续比较时报错
	return -2
end

-- 次数耗尽：已验证成功（cnt=-1）或正常扣减至 0
if cnt <= 0 then
	return -1
end

-- 验证码比对（Lua == 对字符串是值比较，安全）
if inputCode == storedCode then
	-- 验证成功，将次数标记为 -1（语义：「已使用，不可再用」）
	local ttl = redis.call("ttl", cntKey)
	redis.call("set", cntKey, -1)
	if ttl > 0 then
		-- 保持原有的 TTL，让 -1 标记自然随 key 过期而消失
		redis.call("expire", cntKey, ttl)
	end
	return 0
end

-- 验证码错误：扣减一次剩余次数
redis.call("decrby", cntKey, 1)
return -2
