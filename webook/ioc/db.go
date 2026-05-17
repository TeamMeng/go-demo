package ioc

import (
	"fmt"

	"github.com/TeamMeng/go-demo/webook/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("db", &cfg); err != nil {
		panic(err)
	}
	if cfg.DSN == "" {
		panic("missing config: db.dsn")
	} else {
		fmt.Println(cfg.DSN)
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		panic(err)
	}

	if err := dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}
