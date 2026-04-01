// Package lock 的测试文件。
//
// 使用 miniredis 作为内存中的 Redis 服务器进行测试，
// 无需依赖外部的 Redis 实例。
package lock

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// setupTestRedis 创建一个内存中的 Redis 服务器并返回客户端和清理函数。
//
// 参数:
//   - t: 测试实例
//
// 返回:
//   - *redis.Client: Redis 客户端实例
//   - func(): 清理函数，用于关闭客户端和服务器
func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	t.Helper()

	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cleanup := func() {
		client.Close()
		s.Close()
	}

	return client, cleanup
}

// TestTryLock 测试 TryLock 的基本行为。
//
// 验证点:
//   - TryLock 使用 SET 命令设置 key，无论 key 是否存在都返回 true。
//   - 设置后 key 存在且带有过期时间。
func TestTryLock(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()

	key := "test:lock:try"
	expire := 1 * time.Hour

	if !TryLock(client, key, expire) {
		t.Fatal("expected TryLock to succeed on first attempt")
	}

	// 验证 key 存在且有过期时间
	ttl, err := client.TTL(context.Background(), key).Result()
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl <= 0 || ttl > expire {
		t.Fatalf("unexpected TTL: %v", ttl)
	}

	// 当前实现使用普通 SET，会覆盖已有 key，因此第二次也会成功
	if !TryLock(client, key, expire) {
		t.Fatal("expected TryLock to succeed on second attempt (SET overwrites existing key)")
	}
}

// TestReleaseLock 测试 ReleaseLock 的基本行为。
//
// 验证点:
//   - ReleaseLock 能够删除已存在的锁 key。
//   - 释放锁后，TryLock 可以再次成功。
func TestReleaseLock(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()

	key := "test:lock:release"

	if !TryLock(client, key, time.Hour) {
		t.Fatal("expected TryLock to succeed")
	}

	ReleaseLock(client, key)

	exists, err := client.Exists(context.Background(), key).Result()
	if err != nil {
		t.Fatalf("EXISTS failed: %v", err)
	}
	if exists != 0 {
		t.Fatalf("expected key to be deleted, exists=%d", exists)
	}

	if !TryLock(client, key, time.Hour) {
		t.Fatal("expected TryLock to succeed after release")
	}
}

// TestLockRace 测试 LockRace 的并发行为。
//
// 验证点:
//   - 100 个并发协程竞争同一把锁，由于当前实现使用普通 SET，所有协程都能成功设置 key。
//   - 锁在函数结束后被正确释放。
func TestLockRace(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()

	key := "秒杀"

	client.Del(context.Background(), key)

	LockRace(client)

	exists, err := client.Exists(context.Background(), key).Result()
	if err != nil {
		t.Fatalf("EXISTS failed: %v", err)
	}
	if exists != 0 {
		t.Fatalf("expected lock to be released after LockRace, exists=%d", exists)
	}
}

// TestTryLockConcurrency 测试并发场景下 TryLock 的行为。
//
// 验证点:
//   - 当前实现使用普通 SET，多个并发协程同时调用 TryLock 都会成功。
func TestTryLockConcurrency(t *testing.T) {
	t.Parallel()

	client, cleanup := setupTestRedis(t)
	defer cleanup()
	defer client.Del(context.Background(), "test:lock:concurrent")

	const workers = 50
	var successCount int32

	key := "test:lock:concurrent"

	done := make(chan struct{}, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			if TryLock(client, key, time.Hour) {
				atomic.AddInt32(&successCount, 1)
				time.Sleep(10 * time.Millisecond)
				ReleaseLock(client, key)
			}
		}()
	}

	for i := 0; i < workers; i++ {
		<-done
	}

	// 当前实现使用普通 SET，所有协程都能成功
	if atomic.LoadInt32(&successCount) != workers {
		t.Fatalf("expected all %d goroutines to succeed, got %d", workers, successCount)
	}
}
