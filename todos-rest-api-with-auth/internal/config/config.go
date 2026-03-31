// package config 负责从环境变量（及可选的 .env 文件）加载应用配置。
//
// 该包屏蔽了配置来源的复杂性，统一向上层提供结构化的 Config 对象。
// 在本地开发环境中，可以通过项目根目录的 .env 文件设置配置；
// 在生产环境或 CI 中，可以直接使用系统环境变量，无需 .env 文件。
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 保存应用运行所需的核心配置项。
//
// 所有字段均为字符串类型，由调用方根据需要进行类型转换（如端口转整数）。
type Config struct {
	// DatabaseURL 是 PostgreSQL 的连接字符串，格式通常为：
	// postgres://user:password@host:port/dbname?sslmode=disable
	DatabaseURL string

	// Port 是 HTTP 服务监听的端口号，例如 "8080"。
	Port string

	// JWTSecret 是用于签发和校验 JWT Token 的对称密钥。
	// 在生产环境中应使用足够长且随机的字符串，并妥善保管。
	JWTSecret string
}

// Load 从 .env 文件（如果存在）及环境变量中读取配置并返回 Config 实例。
//
// 执行逻辑：
//  1. 尝试调用 godotenv.Load() 加载 .env 文件。
//     若文件不存在，仅打印警告日志，不会返回错误，以便兼容纯环境变量部署。
//  2. 从 os.Getenv 读取 DATABASE_URL、PORT、JWT_SECRET。
//  3. 组装并返回 Config 指针。
//
// 注意：该函数不会校验字段是否为空，调用方（通常是 main）应自行检查必填项。
func Load() (*Config, error) {
	var err error = godotenv.Load()

	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}
	var config *Config = &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	return config, nil
}
