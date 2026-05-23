package main

import (
	// "fmt"
	//
	// "github.com/fsnotify/fsnotify"
	// "github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

// "net/http"

// "github.com/TeamMeng/go-demo/webook/internal/web"
// "github.com/alicebob/miniredis/v2/server"
// "github.com/gin-gonic/gin"

func main() {
	// server := web.InitWe()
	// server := gin.Default()
	//
	// server.GET("/hello", func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusOK, gin.H{"message": "Hello World"})
	// })
	//
	// initViperRemote()
	initViper()
	initLogger()
	server := InitWebServer()

	server.Run(":8080")
}

func initViper() {
	// viper.SetDefault("db.mysql.dsn", "root:root@tcp(localhost:3306)/webook")
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	zap.L().Info("Init successflully")
}

// func initViperRemote() {
// 	if err := viper.AddRemoteProvider("etcd3", "127.0.0.1:2379", "/webook"); err != nil {
// 		panic(err)
// 	}
// 	viper.SetConfigType("yaml")
// 	if err := viper.ReadRemoteConfig(); err != nil {
// 		panic(err)
// 	}
// }

// func initViperWatch() {
// 	cfile := pflag.String("config", "config/config.yaml", "Configuration file path")
// 	pflag.Parse()
// 	viper.SetConfigFile(*cfile)
//
// 	viper.WatchConfig()
// 	viper.OnConfigChange(func(in fsnotify.Event) {
// 		fmt.Println(in.Name, in.Op)
// 	})
//
// 	if err := viper.ReadInConfig(); err != nil {
// 		panic(err)
// 	}
// }
