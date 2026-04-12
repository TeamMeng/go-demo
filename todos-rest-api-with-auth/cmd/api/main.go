// package main 是 Todo API 服务的入口包。
//
// 该包负责整个应用的生命周期管理，包括：
//  1. 加载环境配置（.env 文件及环境变量）
//  2. 建立并验证 PostgreSQL 数据库连接
//  3. 初始化 Gin HTTP 路由引擎
//  4. 注册公开路由（健康检查、注册、登录）
//  5. 注册受保护路由（待办事项的增删改查），并挂载 JWT 认证中间件
//  6. 启动 HTTP 服务器并监听指定端口
//
// 若配置加载或数据库连接失败，程序将直接调用 log.Fatal 终止运行。
package main

import (
	"log"
	"net/http"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/database"
	"todo_api/internal/handlers"
	"todo_api/internal/middleware"
	"todo_api/internal/middleware/ratelimit"
	"todo_api/internal/repository"
	"todo_api/internal/repository/cache"
	"todo_api/internal/service"
	"todo_api/internal/service/sms/aliyun"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// main 是应用的入口函数。
//
// 执行流程如下：
//  1. 调用 config.Load() 读取 DATABASE_URL、PORT、JWT_SECRET 等配置。
//  2. 调用 database.Connect() 创建 pgx 连接池，验证数据库可达性。
//  3. 使用 gin.Default() 创建路由引擎，自动集成 Logger 和 Recovery 中间件。
//  4. 注册根路径 "/" 的健康检查端点。
//  5. 注册认证相关路由：POST /auth/register、POST /auth/login。
//  6. 创建 "/todos" 路由组，挂载 middleware.AuthMiddleware 进行 JWT 校验，
//     并在该组内注册待办事项的 CRUD 路由。
//  7. 注册测试用受保护路由 GET /protected-test。
//  8. 调用 router.Run() 阻塞式启动 HTTP 服务。
//
// 注意：pool 在 main 返回前通过 defer pool.Close() 关闭，确保连接资源释放。
func main() {
	// 加载应用配置，若失败则直接退出
	var cfg *config.Config
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	codeCache := cache.NewCodeCache(redisClient)
	codeRepo := repository.NewCodeRepository(codeCache)

	aliyunClient, err := aliyun.CreateClient()

	if err != nil {
		log.Fatal("Failed to create aliyun sms client:", err)
	}

	smsSvc := aliyun.NewService(tea.String("云渚科技验证平台"), aliyunClient)
	codeSvc := service.NewCodeService(codeRepo, smsSvc)

	// 建立数据库连接池，若失败则直接退出
	var pool *pgxpool.Pool
	pool, err = database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// 确保程序退出时关闭数据库连接池，防止连接泄漏
	defer pool.Close()

	// 初始化 Gin 路由引擎（包含默认的 Logger 和 Recovery 中间件）
	var router *gin.Engine = gin.Default()

	// 禁用代理信任，避免 X-Forwarded-For 等头被恶意利用
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal("Failed to set trusted proxies:", err)
	}

	// 健康检查端点：返回服务运行状态及数据库连接状态
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message":  "Todo API is running",
			"status":   "success",
			"database": "connected",
		})
	})

	// 公开路由：用户注册与登录，无需认证
	router.POST("/auth/register", handlers.CreateUserHandler(pool))
	router.POST("/auth/login", ratelimit.NewBuilder(ratelimit.NewRedisSlidingWindowLimiter(redisClient, time.Minute, 5)).Build(), handlers.LoginHandler(pool, cfg))
	router.POST("/sms/login/code", handlers.SendLoginSMSCodeHandler(codeSvc))
	router.POST("/sms/login/verify", handlers.VerifyLoginSMSCodeHandler(pool, cfg, codeSvc))

	// 受保护路由：所有 /todos 下的接口都需要有效的 JWT Token
	protected := router.Group("/todos")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.POST("", handlers.CreateTodoHandler(pool))
		protected.GET("", handlers.GetAllTodosHandler(pool))
		protected.GET("/:id", handlers.GetToDoByIDHandler(pool, redisClient))
		protected.PUT("/:id", handlers.UpdateTodoHandler(pool))
		protected.DELETE("/:id", handlers.DeleteTodoHandler(pool))
	}

	// 测试用受保护路由，仅用于验证 JWT 中间件是否正常工作
	router.GET("/protected-test", middleware.AuthMiddleware(cfg))

	// 启动 HTTP 服务器，监听 cfg.Port 指定的端口
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
