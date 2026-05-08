//go:build wireinject

package main

import (
	"github.com/TeamMeng/go-demo/webook/internal/repository"
	"github.com/TeamMeng/go-demo/webook/internal/repository/cache"
	"github.com/TeamMeng/go-demo/webook/internal/repository/dao"
	"github.com/TeamMeng/go-demo/webook/internal/service"
	"github.com/TeamMeng/go-demo/webook/internal/web"
	"github.com/TeamMeng/go-demo/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,

		dao.NewUserDAO,

		cache.NewUserCache,
		cache.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,

		ioc.InitSMSService,

		web.NewUserHandler,

		ioc.InitGin,

		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
