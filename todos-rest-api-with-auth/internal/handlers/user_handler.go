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
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/models"
	"todo_api/internal/repository"
	"todo_api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest 是用户注册请求的请求体结构。
//
// 字段说明：
//   - Email: 用户邮箱，必填，作为登录账号。
//   - Password: 用户密码，必填，明文传入，服务端会使用 bcrypt 进行哈希后存储。
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest 是用户登录请求的请求体结构。
//
// 字段说明：
//   - Email: 用户邮箱，必填。
//   - Password: 用户密码，必填，用于与数据库中存储的哈希密码进行比对。
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SmsLoginRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// LoginResponse 是用户登录成功后的响应体结构。
//
// 字段说明：
//   - Token: JWT 访问令牌，客户端应在后续受保护请求的 Authorization 头中以 Bearer 方式携带。
type LoginResponse struct {
	Token string `json:"token"`
}

// CreateUserHandler 返回一个 Gin HandlerFunc，用于处理用户注册请求。
//
// 路由：POST /auth/register（公开）
//
// 处理流程：
//  1. 绑定并校验请求 JSON，确保 email 和 password 字段存在。
//  2. 校验密码长度是否不少于 6 个字符。
//  3. 使用 bcrypt.GenerateFromPassword 对明文密码进行哈希处理。
//  4. 构造 models.User 对象并调用 repository.CreateUser 写入数据库。
//  5. 若邮箱已存在（唯一约束冲突），返回 400 及友好提示；其他错误返回 500。
//  6. 注册成功返回 200 及创建的用户信息（注意：User 结构体的 Password 字段已被 json:"-" 隐藏）。
//
// 参数：
//   - pool: PostgreSQL 连接池，通过闭包传递给 Handler 使用。
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
			Email:    req.Email,
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
func LoginHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
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
		claims := jwt.MapClaims{
			"user_id":    user.ID,
			"email":      user.Email,
			"exp":        time.Now().Add(24 * time.Hour).Unix(),
			"user_agent": ctx.Request.UserAgent(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// 使用配置的 JWTSecret 对 Token 签名
		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token" + err.Error()})
		}

		ctx.JSON(http.StatusOK, &LoginResponse{Token: tokenString})
	}
}

// TestProtectHandler 返回一个 Gin HandlerFunc，用于测试受保护路由的访问。
//
// 路由：通常挂载在需要 AuthMiddleware 之后的路由上。
//
// 处理流程：
//  1. 从上下文中读取 middleware 注入的 user_id。
//  2. 若存在则返回 200 及成功消息和 user_id；若不存在返回 500。
//
// 注意：该 Handler 本身不执行任何业务操作，仅用于验证认证链路是否通畅。
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

func SendLoginSMSCodeHandler(codeSvc *service.CodeService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		const biz = "login"
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
