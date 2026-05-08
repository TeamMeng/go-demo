package main

import (
// "net/http"

// "github.com/TeamMeng/go-demo/webook/internal/web"
// "github.com/alicebob/miniredis/v2/server"
// "github.com/gin-gonic/gin"
)

func main() {
	// server := web.InitWe()
	// server := gin.Default()
	//
	// server.GET("/hello", func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusOK, gin.H{"message": "Hello World"})
	// })
	//
	server := InitWebServer()

	server.Run(":8080")
}
