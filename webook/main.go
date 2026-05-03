package main

import (
	"github.com/TeamMeng/go-demo/webook/internal/web"
)

func main() {
	server := web.InitWeb()

	server.Run(":8080")
}
