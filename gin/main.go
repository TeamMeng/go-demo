package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	server.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	server.POST("/post", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "post",
		})
	})

	server.GET("/users/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		ctx.JSON(http.StatusOK, gin.H{
			"user": name,
		})
	})

	server.GET("/views/*.html", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "This is File Path",
		})
	})

	server.GET("/order", func(ctx *gin.Context) {
		oid := ctx.Query("id")
		ctx.JSON(http.StatusOK, gin.H{
			"order_id": oid,
		})
	})

	server.Run(":8080")
}
