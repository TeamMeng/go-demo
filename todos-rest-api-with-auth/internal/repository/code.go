package repository

import (
	"context"

	"todo_api/internal/repository/cache"
)

var (
	// ErrCodeTooMany 表示验证码发送过于频繁。
	//
	// repository 层直接导出 cache 层错误，方便 service/handler 只依赖 repository 包，
	// 不需要越过 repository 去感知 Redis cache 的具体实现。
	ErrCodeTooMany = cache.ErrCodeTooMany
	// ErrCodeVerifyTooManyTimes 表示同一个验证码的可验证次数已经耗尽。
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

// CodeRepository 是验证码数据访问层。
//
// 目前验证码只存储在 Redis 中，因此 repository 只是对 CodeCache 做一层封装。
// 保留这一层的价值是隔离 service 层和 cache 层：后续如果验证码需要落库、记录审计日志、
// 或切换存储实现，只需要调整 repository，不需要改动业务编排逻辑。
type CodeRepository struct {
	// cache 负责真正的 Redis key 构造、Lua 脚本执行和返回码转换。
	cache *cache.CodeCache
}

// NewCodeRepository 创建验证码 repository。
func NewCodeRepository(c *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: c,
	}
}

// Store 保存验证码。
//
// 参数说明：
//   - biz：业务场景，用于区分不同验证码用途。
//   - phone：手机号，用于定位具体用户的验证码。
//   - code：生成后的验证码明文。
//
// 底层 cache.Set 会通过 Lua 脚本原子化完成“检查发送频率 + 写入验证码 + 初始化验证次数”。
func (repo *CodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

// Verify 校验验证码。
//
// 返回 true 表示验证码匹配；返回 false 表示验证码不匹配或无法通过验证。
// 具体的次数耗尽、未知状态、Redis 异常等情况会通过 error 暴露给上层。
func (repo *CodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, code)
}
