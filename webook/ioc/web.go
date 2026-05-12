package ioc

import (
	"strings"
	"time"

	"github.com/TeamMeng/go-demo/webook/internal/web"
	"github.com/TeamMeng/go-demo/webook/internal/web/middleware"
	"github.com/TeamMeng/go-demo/webook/pkg/ginx/middlewares/ratelimit"
	pkgratelimit "github.com/TeamMeng/go-demo/webook/pkg/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitGin(mdls []gin.HandlerFunc, hld *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hld.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),

		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").
			IgnorePaths("/users/loginJWT").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/users/refresh_token").
			Build(),

		ratelimit.NewBuilder(pkgratelimit.NewRedisSlidingWindowLimiter(redisClient, time.Minute, 200)).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	})
}
