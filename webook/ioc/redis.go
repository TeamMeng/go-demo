package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type Config struct {
		ADDR string `yaml:"addr"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("redis", &cfg); err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ADDR,
	})
	return redisClient
}
