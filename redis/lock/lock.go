// Package lock 提供了基于 Redis 的简单锁实现。
//
// 当前实现使用 Redis 的 SET 命令来设置锁，使用 DEL 命令来释放锁。
// 注意：此实现不是分布式锁的完整解决方案，仅用于演示目的。
// 在生产环境中，建议使用 Redlock 算法或其他成熟的分布式锁实现。
package lock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// TryLock 尝试获取一个基于 Redis 的锁。
//
// 参数:
//   - rc: Redis 客户端实例
//   - key: 锁的键名
//   - expire: 锁的过期时间
//
// 返回:
//   - bool: 如果成功设置 key 则返回 true，否则返回 false
//
// 注意: 当前实现使用普通的 SET 命令，会覆盖已存在的 key。
// 这意味着即使锁已经被其他客户端持有，调用此方法也会成功。
func TryLock(rc *redis.Client, key string, expire time.Duration) bool {
	cmd := rc.Set(context.Background(), key, "value", expire)
	return cmd.Err() == nil
}

// ReleaseLock 释放一个基于 Redis 的锁。
//
// 参数:
//   - rc: Redis 客户端实例
//   - key: 锁的键名
//
// 注意: 此方法会删除指定的 key，无论当前客户端是否持有该锁。
// 在生产环境中，应该使用 Lua 脚本来确保只有锁的持有者才能释放锁。
func ReleaseLock(rc *redis.Client, key string) {
	rc.Del(context.Background(), key)
}

// LockRace 演示多个协程竞争同一把锁的场景。
//
// 参数:
//   - client: Redis 客户端实例
//
// 该方法启动 100 个并发协程，每个协程都尝试获取名为"秒杀"的锁。
// 由于当前实现使用普通 SET 命令，所有协程都能成功获取锁。
// 函数结束时，锁会被释放。
//
// 注意: 此函数仅用于演示和测试目的。
func LockRace(client *redis.Client) {
	key := "秒杀"
	defer ReleaseLock(client, key)
	const P = 100
	wg := sync.WaitGroup{}
	wg.Add(P)
	for i := 0; i < P; i++ {
		go func(i int) {
			defer wg.Done()
			if TryLock(client, key, time.Hour) {
				fmt.Printf("协程%d抢到锁\n", i)
			}
		}(i)
	}
	wg.Wait()
}
