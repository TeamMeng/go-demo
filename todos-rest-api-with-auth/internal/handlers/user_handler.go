// Package handlers 包含所有 HTTP 请求处理器（Handler），负责解析请求、调用仓库层并返回响应。
//
// 该层是应用的接入层（Transport Layer），主要职责包括：
//  1. 解析并校验请求参数（JSON 请求体）。
//  2. 调用 repository 层完成用户注册、登录等业务操作。
//  3. 使用 bcrypt 进行密码哈希和校验。
//  4. 使用 JWT 签发认证 Token。
//  5. 根据操作结果返回对应的 HTTP 状态码和 JSON 响应。
package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/models"
	"todo_api/internal/repository"
	"todo_api/internal/service"
	jwtpkg "todo_api/internal/service/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const biz = "login"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SmsLoginRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type SmsLoginVerifyRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

func CreateUserHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "jsonReq" + err.Error()})
			return
		}

		// 密码长度校验
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
			return
		}

		// 使用 bcrypt 对密码进行哈希，DefaultCost 为默认计算成本（当前为 10）
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password" + err.Error()})
			return
		}

		user := &models.User{
			Email:    sql.NullString{String: req.Email, Valid: true},
			Password: string(hashedPassword),
		}

		createdUser, err := repository.CreateUser(pool, user)
		if err != nil {
			// 通过错误信息关键字判断是否为邮箱唯一约束冲突
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user" + err.Error()})
			return
		}

		c.JSON(http.StatusOK, createdUser)
	}
}

// LoginHandler 返回一个 Gin HandlerFunc，用于处理用户登录请求并签发 JWT Token。
//
// 路由：POST /auth/login（公开）
//
// 处理流程：
//  1. 绑定并校验请求 JSON，提取邮箱和密码。
//  2. 调用 repository.GetUserByEmail 查询用户信息；若查询失败（用户不存在），返回 401。
//  3. 使用 bcrypt.CompareHashAndPassword 比对明文密码与数据库中的哈希密码；不匹配返回 401。
//  4. 构造 JWT Claims，包含 user_id、email 和 24 小时后的过期时间（exp）。
//  5. 使用 HS256 算法和 cfg.JWTSecret 对 Token 进行签名。
//  6. 成功返回 200 及包含 Token 的 LoginResponse。
//
// 参数：
//   - pool: PostgreSQL 连接池。
//   - cfg: 应用配置，用于获取 JWT 签名密钥。
func LoginHandler(pool *pgxpool.Pool, cfg *config.Config, jwtSvc *jwtpkg.JWTHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req LoginRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "jsonReq" + err.Error()})
			return
		}

		// 根据邮箱查询用户，若用户不存在则返回统一的"Invalid credentials"，避免枚举攻击
		user, err := repository.GetUserByEmail(pool, req.Email)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// 校验密码，若错误也返回统一的"Invalid credentials"
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// 构造 JWT Claims，设置 24 小时有效期
		// 注意：额外添加 user_agent 字段，将 Token 绑定到签发时的客户端 User-Agent。
		// 若他人盗取 Token 并在不同 User-Agent 下使用，AuthMiddleware 会拒绝该请求，提升安全性。

		// 使用配置的 JWTSecret 对 Token 签名
		if err = jwtSvc.SetJwtToken(ctx, user.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate accessed token" + err.Error()})
		}

		if err = jwtSvc.SetRefreshToken(ctx, user.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token" + err.Error()})
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "successfully"})
	}
}

func TestProtectHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, exists := ctx.Get("user_id")

		if !exists {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in context"})
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Protected route accessed successfully",
			"user_id": userID,
		})
	}
}

// SendLoginSMSCodeHandler 返回一个 Gin HandlerFunc，用于向用户手机发送登录验证码。
//
// 路由：POST /auth/login/sms（公开）
//
// 处理流程：
//  1. 绑定并校验请求 JSON，确保 phone 字段存在。
//  2. 调用 CodeService.Send 生成 6 位验证码、写入 Redis 并发送短信。
//  3. 发送失败时返回 500，由 service 层控制发送频率（540 秒冷却期）。
//  4. 成功返回 200。
//
// 安全说明：
//
//	发送频率由 Lua 脚本在 Redis 层面控制，同一手机号 540 秒内只能获取一次验证码。
//	避免短信轰炸攻击。
func SendLoginSMSCodeHandler(codeSvc *service.CodeService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SmsLoginRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "jsonReq" + err.Error()})
			return
		}
		if err := codeSvc.Send(ctx, biz, req.Phone); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Send successfully",
		})
	}
}

// VerifyLoginSMSCodeHandler 返回一个 Gin HandlerFunc，用于验证短信验证码并完成登录。
//
// 路由：POST /auth/login/sms/verify（公开）
//
// 处理流程：
//  1. 绑定并校验请求 JSON，提取 phone 和 code。
//  2. 调用 CodeService.Verify 校验验证码：
//     - 验证成功：Lua 脚本自动将验证次数置为 -1，防止验证码重复使用。
//     - 验证失败：返回 401。
//  3. 验证码通过后，调用 repository.FindOrCreateUserByPhone 查找或创建账户。
//  4. 签发 JWT Token（24 小时有效），并绑定 User-Agent。
//  5. 返回 Token 给客户端。
//
// 安全说明：
//   - 验证码通过后立即使用，不支持「先验证再决定是否登录」的犹豫期。
//   - JWT 绑定 User-Agent，防止 Token 被盗后在其他客户端使用。
func VerifyLoginSMSCodeHandler(pool *pgxpool.Pool, cfg *config.Config, codeSvc *service.CodeService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SmsLoginVerifyRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "jsonReq" + err.Error()})
			return
		}
		ok, err := codeSvc.Verify(ctx, biz, req.Phone, req.Code)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
			return
		}

		user, err := repository.FindOrCreateUserByPhone(pool, req.Phone)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 构造 JWT Claims，设置 24 小时有效期
		// 注意：额外添加 user_agent 字段，将 Token 绑定到签发时的客户端 User-Agent。
		// 若他人盗取 Token 并在不同 User-Agent 下使用，AuthMiddleware 会拒绝该请求，提升安全性。
		claims := jwt.MapClaims{
			"user_id":    user.ID,
			"phone":      user.Phone,
			"exp":        time.Now().Add(24 * time.Hour).Unix(),
			"user_agent": ctx.Request.UserAgent(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// 使用配置的 JWTSecret 对 Token 签名
		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token" + err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Login successfully",
			"token":   tokenString,
		})
	}
}
