package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// server := web.InitWeb()
	server := gin.Default()

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Hello World"})
	})

	server.Run(":8080")
}
