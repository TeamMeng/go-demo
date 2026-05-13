package web

import (
	"strings"
	"time"

	"github.com/TeamMeng/go-demo/webook/internal/repository"
	"github.com/TeamMeng/go-demo/webook/internal/repository/cache"
	"github.com/TeamMeng/go-demo/webook/internal/repository/dao"
	"github.com/TeamMeng/go-demo/webook/internal/service"
	"github.com/TeamMeng/go-demo/webook/internal/service/sms/memory"
	"github.com/TeamMeng/go-demo/webook/internal/web/jwt"
	"github.com/TeamMeng/go-demo/webook/internal/web/middleware"
	"github.com/TeamMeng/go-demo/webook/pkg/ginx/middlewares/ratelimit"
	pkgratelimit "github.com/TeamMeng/go-demo/webook/pkg/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/redis/go-redis/v9"

	// "github.com/gin-contrib/sessions"
	// "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitWeb() *gin.Engine {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	db := initDB()
	initUser(server, db)

	return server
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/webook"))
	if err != nil {
		panic(err)
	}

	if err := dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}

func initUser(server *gin.Engine, db *gorm.DB) {
	// store := cookie.NewStore([]byte("secret"))
	// store := memstore.NewStore([]byte("EP4UNRkorqfjiac2bt6CVH1QuCEYlISP"),
	// 	[]byte("FQMsL9km5fD92HuBwsCnG9f6JWuFKGhA"))

	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "", "",
	// 	[]byte("EP4UNRkorqfjiac2bt6CVH1QuCEYlISP"),
	// 	[]byte("FQMsL9km5fD92HuBwsCnG9f6JWuFKGhA"))
	// if err != nil {
	// 	panic(err)
	// }

	// server.Use(sessions.Sessions("mysession", store))
	//
	// server.Use(middleware.NewLoginMiddlewareBuilder().
	// 	IgnorePaths("/users/signup").
	// 	IgnorePaths("/users/login").
	// 	IgnorePaths("/users/loginJWT").
	// 	Build(),
	// )

	redisClient := initRedis()
	jwtHandler := jwt.NewRedisJWT(redisClient)

	server.Use(ratelimit.NewBuilder(pkgratelimit.NewRedisSlidingWindowLimiter(redisClient, time.Minute, 200)).Build())

	server.Use(middleware.NewLoginJWTMiddlewareBuilder(jwtHandler).
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		IgnorePaths("/users/loginJWT").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").
		IgnorePaths("/users/refresh_token").
		Build(),
	)

	codeCache := cache.NewCodeCache(redisClient)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()

	cache := cache.NewUserCache(redisClient)
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud, cache)

	svc := service.NewUserService(repo)
	codeSvc := service.NewCodeService(codeRepo, smsSvc)
	u := NewUserHandler(svc, codeSvc, jwtHandler)
	u.RegisterRoutes(server)
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return redisClient
}
