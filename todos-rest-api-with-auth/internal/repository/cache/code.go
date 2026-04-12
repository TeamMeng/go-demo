package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrCodeTooMany 表示当前验证码还在发送冷却期内，调用方应该提示用户稍后再试。
	ErrCodeTooMany = errors.New("code is too many")
	// ErrCodeVerifyTooManyTimes 表示验证码验证次数已经用完，调用方应该要求用户重新获取验证码。
	ErrCodeVerifyTooManyTimes = errors.New("code verify is too many")
	// ErrUnknownForCode 表示 Lua 脚本返回了 cache 层没有明确识别的状态码。
	// 这通常代表验证码不存在、脚本返回码和 Go 代码没有对齐，或 Redis 数据被外部修改。
	ErrUnknownForCode = errors.New("unknown code")
)

// luaSetCode 是嵌入到 Go 二进制里的验证码写入脚本。
//
// 使用 Lua 的原因是要把“读取 TTL、判断发送频率、写验证码、写验证次数”放到 Redis
// 内部一次性执行，避免并发请求下出现先检查后写入的竞态问题。
//
//go:embed lua/set_code.lua
var luaSetCode string

// luaVerifyCode 是嵌入到 Go 二进制里的验证码校验脚本。
//
// 脚本会原子化完成“读取验证码、读取剩余次数、比对验证码、更新次数”。
//
//go:embed lua/verify_code.lua
var luaVerifyCode string

// CodeCache 封装验证码在 Redis 中的读写逻辑。
//
// Redis 中使用两个 key：
//   - phone_code:{biz}:{phone}：验证码本体。
//   - phone_code:{biz}:{phone}:cnt：剩余可验证次数。
//
// 两个 key 由同一组 Lua 脚本维护，避免验证码和次数的状态不一致。
type CodeCache struct {
	// client 使用 redis.Cmdable，便于在测试中传入 redis.Client、redis.ClusterClient 或 mock。
	client redis.Cmdable
}

// NewCodeCache 创建验证码缓存对象。
func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}
}

// Set 写入验证码并初始化验证次数。
//
// 这里不直接调用 SET/EXPIRE，而是执行 luaSetCode：
//   - 如果验证码不存在，或剩余有效期小于 9 分钟，则允许重新发送并重置 10 分钟有效期。
//   - 如果验证码剩余有效期仍然较长，则拒绝发送，用于限制频繁获取验证码。
//   - 每次成功写入验证码时，会把验证次数初始化为 5 次。
//
// Lua 返回码约定：
//   - 0：写入成功。
//   - -1：发送过于频繁。
//   - 其它：Go 层视为系统错误。
func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1:
		return ErrCodeTooMany
	default:
		return errors.New("system error")
	}
}

// Verify 校验验证码并维护剩余验证次数。
//
// luaVerifyCode 会在 Redis 内部一次性完成校验和次数更新：
//   - 验证成功时返回 0，并把次数标记为 -1，避免验证码被重复使用。
//   - 验证次数耗尽时返回 -1。
//   - 验证码错误或验证码已过期时返回 -2。
//   - 其它状态会被映射成 ErrUnknownForCode。
//
// 返回值中 bool 只表达“验证码是否正确”，error 用于表达限流、次数耗尽或系统异常。
func (c *CodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	default:
		return false, ErrUnknownForCode
	}
}

// key 生成验证码在 Redis 中的主 key。
//
// biz 作为 key 的一部分，可以让同一个手机号在不同业务场景下拥有独立验证码。
func (c *CodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
