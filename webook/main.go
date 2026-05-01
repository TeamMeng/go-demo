package main

import (
	"github.com/TeamMeng/go-demo/webook/internal/web"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	u := &web.UserHandler{}
	u.RegisterRoutes(server)

	server.Run(":8080")
}
