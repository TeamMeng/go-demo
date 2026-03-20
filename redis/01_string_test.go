package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisString(t *testing.T) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer rdb.Close()

	// 教学场景下，如果本机没有启动 Redis，就跳过测试。
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis is not available: %v", err)
	}

	t.Run("1. 基础读写", func(t *testing.T) {
		key := "user:1001:name"
		rdb.Del(ctx, key)

		// SET: 写入一个字符串
		if err := rdb.Set(ctx, key, "Alice", 0).Err(); err != nil {
			t.Fatalf("SET failed: %v", err)
		}

		// GET: 读取字符串
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			t.Fatalf("GET failed: %v", err)
		}
		if val != "Alice" {
			t.Fatalf("GET = %q, want %q", val, "Alice")
		}

		// DEL: 删除后再读，应该得到 redis.Nil
		if err := rdb.Del(ctx, key).Err(); err != nil {
			t.Fatalf("DEL failed: %v", err)
		}

		_, err = rdb.Get(ctx, key).Result()
		if !errors.Is(err, redis.Nil) {
			t.Fatalf("GET after DEL error = %v, want redis.Nil", err)
		}

		// 重新写入新值
		if err := rdb.Set(ctx, key, "Bob", 0).Err(); err != nil {
			t.Fatalf("SET new value failed: %v", err)
		}

		val, err = rdb.Get(ctx, key).Result()
		if err != nil {
			t.Fatalf("GET new value failed: %v", err)
		}
		if val != "Bob" {
			t.Fatalf("GET new value = %q, want %q", val, "Bob")
		}
	})

	t.Run("2. 存储数字", func(t *testing.T) {
		key := "counter"
		rdb.Del(ctx, key)

		// Redis 的 String 也可以存数字，本质仍然是字符串。
		if err := rdb.Set(ctx, key, "100", 0).Err(); err != nil {
			t.Fatalf("SET counter failed: %v", err)
		}

		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			t.Fatalf("GET counter failed: %v", err)
		}
		if val != "100" {
			t.Fatalf("GET counter = %q, want %q", val, "100")
		}
	})

	t.Run("3. 存储 JSON", func(t *testing.T) {
		key := "tokeninfo:1:0x123"
		rdb.Del(ctx, key)

		jsonStr := `{"name":"USDT","symbol":"USDT","decimals":18}`
		if err := rdb.Set(ctx, key, jsonStr, 0).Err(); err != nil {
			t.Fatalf("SET json failed: %v", err)
		}

		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			t.Fatalf("GET json failed: %v", err)
		}
		if val != jsonStr {
			t.Fatalf("GET json = %q, want %q", val, jsonStr)
		}
	})

	t.Run("4. 过期时间", func(t *testing.T) {
		key := "cache:user:1001"
		rdb.Del(ctx, key)

		// 写入时直接附带 TTL。
		if err := rdb.Set(ctx, key, "cached_data", 200*time.Millisecond).Err(); err != nil {
			t.Fatalf("SET with TTL failed: %v", err)
		}

		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			t.Fatalf("GET before expire failed: %v", err)
		}
		if val != "cached_data" {
			t.Fatalf("GET before expire = %q, want %q", val, "cached_data")
		}

		time.Sleep(300 * time.Millisecond)

		_, err = rdb.Get(ctx, key).Result()
		if !errors.Is(err, redis.Nil) {
			t.Fatalf("GET after expire error = %v, want redis.Nil", err)
		}
	})

	t.Run("5. SETNX", func(t *testing.T) {
		key := "lock:order:1001"
		rdb.Del(ctx, key)

		// 第一次写入成功，因为 key 不存在。
		ok, err := rdb.SetNX(ctx, key, "locked", 0).Result()
		if err != nil {
			t.Fatalf("first SETNX failed: %v", err)
		}
		if !ok {
			t.Fatalf("first SETNX = %v, want true", ok)
		}

		// 第二次写入失败，因为 key 已经存在。
		ok, err = rdb.SetNX(ctx, key, "locked", 0).Result()
		if err != nil {
			t.Fatalf("second SETNX failed: %v", err)
		}
		if ok {
			t.Fatalf("second SETNX = %v, want false", ok)
		}

		// 更推荐的写法：一次完成 NX + EX。
		cacheKey := "tokeninfo:1:0x789"
		rdb.Del(ctx, cacheKey)

		reply, err := rdb.SetArgs(ctx, cacheKey, "token_data", redis.SetArgs{
			Mode: "NX",
			TTL:  30 * time.Second,
		}).Result()
		if err != nil {
			t.Fatalf("SET NX EX failed: %v", err)
		}
		if reply != "OK" {
			t.Fatalf("SET NX EX reply = %q, want %q", reply, "OK")
		}
	})

	t.Run("6. 计数器", func(t *testing.T) {
		key := "counter:page:views"
		rdb.Del(ctx, key)

		v1, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			t.Fatalf("INCR 1 failed: %v", err)
		}

		v2, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			t.Fatalf("INCR 2 failed: %v", err)
		}

		v3, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			t.Fatalf("INCR 3 failed: %v", err)
		}

		v4, err := rdb.IncrBy(ctx, key, 5).Result()
		if err != nil {
			t.Fatalf("INCRBY failed: %v", err)
		}

		v5, err := rdb.Decr(ctx, key).Result()
		if err != nil {
			t.Fatalf("DECR failed: %v", err)
		}

		v6, err := rdb.DecrBy(ctx, key, 5).Result()
		if err != nil {
			t.Fatalf("DECRBY failed: %v", err)
		}

		got := []int64{v1, v2, v3, v4, v5, v6}
		want := []int64{1, 2, 3, 8, 7, 2}

		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("step %d = %d, want %d", i+1, got[i], want[i])
			}
		}
	})
}
