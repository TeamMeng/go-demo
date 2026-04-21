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
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/middleware"
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

func RefreshTokenHandler(cfg *config.Config, jwtSvc *jwtpkg.JWTHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refreshToken := middleware.ExtractToken(ctx)
		if refreshToken == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
			// 强制校验签名算法，防止算法切换攻击（alg:none、RS256 等）
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			log.Printf("JWT parse error: %v, token.Valid: %v", err, token != nil && token.Valid)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		Claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token Claims"})
			return
		}

		userID, ok := Claims["user_id"].(string)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token Claims"})
			return
		}

		jwtSvc.SetRefreshToken(ctx, userID)
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Refresh token successfully",
		})
	}
}
