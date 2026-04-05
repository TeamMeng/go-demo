// Package ratelimit 提供限流中间件，用于保护 API 服务免受过度请求。
// 支持自定义限流 key 生成函数和日志函数，便于集成到各种业务场景。
package ratelimit

import (
	"log"
	"net/http"
	"strings"
	"todo_api/internal/ratelimit"

	"github.com/gin-gonic/gin"
)

// Builder 是限流中间件的构建器。
// 通过链式调用配置限流器、key 生成函数和日志函数，最后调用 Build() 生成中间件。
type Builder struct {
	limiter  ratelimit.Limiter             // 限流器实例，实现 ratelimit.Limiter 接口
	genKeyFn func(ctx *gin.Context) string // key 生成函数，用于标识被限流的资源
	logFn    func(msg any, args ...any)    // 日志函数，用于记录限流相关的错误信息
}

// NewBuilder 创建一个新的限流中间件构建器。
// limiter: 限流器实例，必须实现 ratelimit.Limiter 接口。
// 返回: 配置好的 Builder 实例。
func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{
		limiter: limiter,
		// 默认 key 生成函数，使用客户端 IP 地址作为限流 key
		genKeyFn: func(ctx *gin.Context) string {
			var b strings.Builder
			b.WriteString("ip-limiter")
			b.WriteString(":")
			b.WriteString(ctx.ClientIP())
			return b.String()
		},
		// 默认日志函数，使用标准库的 log.Println 输出
		logFn: func(msg any, args ...any) {
			v := make([]any, 0, len(args)+1)
			v = append(v, msg)
			v = append(v, args...)
			log.Println(v...)
		},
	}
}

// SetKeyGenFunc 设置自定义的 key 生成函数。
// fn: 接收 gin.Context 并返回唯一标识请求的字符串（如用户ID、IP、API端点等）。
// 返回 Builder 自身，支持链式调用。
func (b *Builder) SetKeyGenFunc(fn func(*gin.Context) string) *Builder {
	b.genKeyFn = fn
	return b
}

// SetLogFunc 设置自定义的日志函数。
// fn: 接收日志消息和可选参数，用于记录限流错误信息。
// 返回 Builder 自身，支持链式调用。
func (b *Builder) SetLogFunc(fn func(msg any, args ...any)) *Builder {
	b.logFn = fn
	return b
}

// Build 生成并返回 Gin 中间件处理器函数。
// 返回: gin.HandlerFunc，可直接用于 Gin 路由注册。
//
// 中间件处理流程:
//  1. 调用 limit() 检查请求是否应该被限流
//  2. 如果发生错误，返回 500 Internal Server Error
//  3. 如果请求被限流，返回 429 Too Many Requests
//  4. 否则调用 ctx.Next() 继续处理链
func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limiter, err := b.limit(ctx)
		if err != nil {
			b.logFn(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limiter {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

// limit 调用限流器判断请求是否应该被限流。
// ctx: Gin 上下文，包含请求信息。
// 返回: (是否限流, 错误)。
//   - (true, nil): 请求被限流
//   - (false, nil): 请求允许通过
//   - (_, error): 发生错误
func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	return b.limiter.Limit(ctx, b.genKeyFn(ctx))
}
