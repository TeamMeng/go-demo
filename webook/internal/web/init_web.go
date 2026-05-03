package web

import (
	"github.com/TeamMeng/go-demo/webook/internal/repository"
	"github.com/TeamMeng/go-demo/webook/internal/repository/dao"
	"github.com/TeamMeng/go-demo/webook/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitWeb() *gin.Engine {
	server := gin.Default()

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
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := NewUserHandler(svc)
	u.RegisterRoutes(server)
}
