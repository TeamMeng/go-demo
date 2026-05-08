//go:build wireinject

// 让 wire 注入
package wire

import (
	"github.com/TeamMeng/go-demo/webook/wire/repository"
	"github.com/TeamMeng/go-demo/webook/wire/repository/dao"
	"github.com/google/wire"
)

func InitRepository() *repository.UserRepository {
	wire.Build(repository.NewUserRepository, dao.NewUserDAO, InitDB)
	return new(repository.UserRepository)
}
