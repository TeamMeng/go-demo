package ioc

import (
	"log"
	"time"

	"github.com/TeamMeng/go-demo/webook/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger.New(log.Default(), logger.Config{
			SlowThreshold:        time.Millisecond * 10,
			Colorful:             true,
			LogLevel:             logger.Info,
			ParameterizedQueries: true,
		}),
	})
	if err != nil {
		panic(err)
	}

	if err := dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}
