// package database 提供数据库连接管理功能。
//
// 该包封装了基于 pgx/v5 的 PostgreSQL 连接池创建逻辑，
// 向上层返回一个可直接用于并发查询的 *pgxpool.Pool 实例。
package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect 解析 databaseURL 并创建 PostgreSQL 连接池，成功后会执行 Ping 验证连通性。
//
// 参数：
//   - databaseURL: PostgreSQL 连接字符串，例如：
//     "postgres://user:password@localhost:5432/todo_db?sslmode=disable"
//
// 返回值：
//   - *pgxpool.Pool: 可用于并发查询的数据库连接池。
//   - error: 若解析连接字符串、创建连接池或 Ping 失败，则返回具体错误。
//
// 执行流程：
//  1. 使用 pgxpool.ParseConfig 解析连接字符串。
//  2. 使用 pgxpool.NewWithConfig 创建连接池。
//  3. 调用 pool.Ping 验证数据库是否可达。
//  4. 若 Ping 失败，立即关闭连接池并返回错误，避免将不可用的池交给上层。
//
// 注意：调用方负责在应用退出时调用 pool.Close() 释放资源。
func Connect(databaseURL string) (*pgxpool.Pool, error) {
	var config *pgxpool.Config
	var err error
	config, err = pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Printf("Unable to parse DATABASE_URL: %v", err)
		return nil, err
	}

	var pool *pgxpool.Pool
	var ctx context.Context = context.Background()
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Printf("Unable to create connection pool: %v", err)
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.Printf("Unable to ping database: %v", err)
		pool.Close()
		return nil, err
	}

	log.Println("Successfully connected to PostgreSQL database")
	return pool, nil
}
