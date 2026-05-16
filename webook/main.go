package main

import "github.com/spf13/viper"

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
	initViper()
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
