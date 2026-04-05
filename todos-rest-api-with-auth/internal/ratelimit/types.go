package ratelimit

import "context"

// Limiter 是限流器的通用接口。
// 所有限流实现都必须实现此接口，以保证不同算法间的可替换性。
type Limiter interface {
	// Limit 判断给定 key 的请求是否应该被限流。
	// ctx: 上下文，用于控制超时和取消。
	// key: 标识被限流资源或用户的唯一键（如用户ID、IP、API端点等）。
	// 返回 (是否限流, 错误)。
	//   - 返回 (true, nil) 表示请求被限流，应拒绝。
	//   - 返回 (false, nil) 表示请求允许通过。
	//   - 返回 (_, error) 表示发生错误，需要根据业务场景决定是限流还是放行。
	Limit(ctx context.Context, key string) (bool, error)
}
