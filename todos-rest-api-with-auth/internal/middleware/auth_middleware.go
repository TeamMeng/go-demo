// package middleware 提供 Gin 中间件，用于请求拦截、认证校验等横切关注点。
//
// 该包中的中间件可在路由组或单个路由上复用，避免在每个 Handler 中重复编写认证逻辑。
package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware 返回一个 JWT 认证中间件（gin.HandlerFunc）。
//
// 该中间件负责校验 HTTP 请求头 Authorization 中携带的 Bearer Token：
//  1. 检查 Authorization 头是否存在；不存在返回 401。
//  2. 提取并校验 Token 格式（必须以 "Bearer " 开头）；格式错误返回 401。
//  3. 使用 jwt.Parse 解析 Token，校验签名算法必须为 HS256，并使用 cfg.JWTSecret 验证签名。
//  4. 若 Token 解析失败或无效，返回 401 及 "Invalid or expired token"。
//  5. 提取 Claims 中的 user_id 并校验类型；失败返回 401。
//  6. 额外校验 exp（过期时间）字段，若 Token 已过期返回 401。
//  7. 将 user_id 写入 Gin 上下文（c.Set("user_id", userID)），供后续 Handler 使用。
//  8. 调用 c.Next() 放行请求，继续执行后续中间件和路由 Handler。
//
// 参数：
//   - cfg: 应用配置，必须包含有效的 JWTSecret 用于签名验证。
//
// 使用示例：
//
//	protected := router.Group("/todos")
//	protected.Use(middleware.AuthMiddleware(cfg))
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 读取 Authorization 请求头
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			ctx.Abort()
			return
		}
		// 去掉 "Bearer " 前缀，获取原始 Token 字符串
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 若去掉前缀后为空或和原字符串相同，说明格式不正确（未以 Bearer 开头）
		if tokenString == "" || tokenString == authHeader {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			ctx.Abort()
			return
		}

		// 解析并验证 JWT Token
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			// 强制校验签名算法，防止算法切换攻击（alg:none、RS256 等）
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			log.Printf("JWT parse error: %v, token.Valid: %v", err, token != nil && token.Valid)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			ctx.Abort()
			return
		}

		// 提取 Claims 并断言为 jwt.MapClaims
		Claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token Claims"})
			ctx.Abort()
			return
		}

		// 从 Claims 中提取 user_id，并确保其为字符串类型
		userID, ok := Claims["user_id"].(string)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token Claims"})
			ctx.Abort()
			return
		}

		// 校验 Token 的过期时间（exp），虽然 jwt.Parse 已做校验，但此处显式校验增加安全性
		if exp, ok := Claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			if time.Now().After(expirationTime) {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
				ctx.Abort()
				return
			}
		}

		// 将 user_id 存入上下文，供下游 Handler 读取
		ctx.Set("user_id", userID)
		ctx.Next()
	}
}
