package main

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedisHash(t *testing.T) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis is not available: %v", err)
	}

	t.Run("1. 基础读写", func(t *testing.T) {
		key := "user:1001"
		rdb.Del(ctx, key)

		if err := rdb.HSet(ctx, key, "name", "Alice").Err(); err != nil {
			t.Fatalf("HSET failed: %v", err)
		}

		val, err := rdb.HGet(ctx, key, "name").Result()
		if err != nil {
			t.Fatalf("HGET failed: %v", err)
		}
		if val != "Alice" {
			t.Fatalf("HGET = %q, want %q", val, "Alice")
		}
	})

	t.Run("2. 一次写入多个字段", func(t *testing.T) {
		key := "user:1002"
		rdb.Del(ctx, key)

		// 一次写入多个 field/value。
		if err := rdb.HSet(ctx, key, map[string]string{
			"name":  "Bob",
			"age":   "30",
			"email": "bob@example.com",
		}).Err(); err != nil {
			t.Fatalf("HSET multiple fields failed: %v", err)
		}

		tests := map[string]string{
			"name":  "Bob",
			"age":   "30",
			"email": "bob@example.com",
		}

		for field, want := range tests {
			got, err := rdb.HGet(ctx, key, field).Result()
			if err != nil {
				t.Fatalf("HGET %s failed: %v", field, err)
			}
			if got != want {
				t.Fatalf("HGET %s = %q, want %q", field, got, want)
			}
		}
	})

	t.Run("3. 读取全部字段", func(t *testing.T) {
		key := "user:1003"
		rdb.Del(ctx, key)

		want := map[string]string{
			"name":  "Carol",
			"age":   "26",
			"email": "carol@example.com",
		}

		if err := rdb.HSet(ctx, key, want).Err(); err != nil {
			t.Fatalf("HSET failed: %v", err)
		}

		got, err := rdb.HGetAll(ctx, key).Result()
		if err != nil {
			t.Fatalf("HGETALL failed: %v", err)
		}

		if len(got) != len(want) {
			t.Fatalf("HGETALL len = %d, want %d", len(got), len(want))
		}

		for field, wantVal := range want {
			if got[field] != wantVal {
				t.Fatalf("HGETALL[%q] = %q, want %q", field, got[field], wantVal)
			}
		}
	})

	t.Run("4. 判断字段是否存在", func(t *testing.T) {
		key := "user:1004"
		rdb.Del(ctx, key)

		if err := rdb.HSet(ctx, key, map[string]string{
			"name": "Dave",
			"age":  "31",
		}).Err(); err != nil {
			t.Fatalf("HSET failed: %v", err)
		}

		exists, err := rdb.HExists(ctx, key, "age").Result()
		if err != nil {
			t.Fatalf("HEXISTS existing field failed: %v", err)
		}
		if !exists {
			t.Fatalf("HEXISTS existing field = %v, want true", exists)
		}

		exists, err = rdb.HExists(ctx, key, "phone").Result()
		if err != nil {
			t.Fatalf("HEXISTS missing field failed: %v", err)
		}
		if exists {
			t.Fatalf("HEXISTS missing field = %v, want false", exists)
		}
	})

	t.Run("5. 删除字段", func(t *testing.T) {
		key := "user:1005"
		rdb.Del(ctx, key)

		if err := rdb.HSet(ctx, key, map[string]string{
			"name":  "Eve",
			"email": "eve@example.com",
		}).Err(); err != nil {
			t.Fatalf("HSET failed: %v", err)
		}

		if err := rdb.HDel(ctx, key, "email").Err(); err != nil {
			t.Fatalf("HDEL failed: %v", err)
		}

		_, err := rdb.HGet(ctx, key, "email").Result()
		if !errors.Is(err, redis.Nil) {
			t.Fatalf("HGET after HDEL error = %v, want redis.Nil", err)
		}

		val, err := rdb.HGet(ctx, key, "name").Result()
		if err != nil {
			t.Fatalf("HGET remaining field failed: %v", err)
		}
		if val != "Eve" {
			t.Fatalf("remaining field = %q, want %q", val, "Eve")
		}
	})

	t.Run("6. 获取字段数量", func(t *testing.T) {
		key := "user:1006"
		rdb.Del(ctx, key)

		if err := rdb.HSet(ctx, key, map[string]string{
			"name":  "Frank",
			"age":   "40",
			"email": "frank@example.com",
		}).Err(); err != nil {
			t.Fatalf("HSET failed: %v", err)
		}

		n, err := rdb.HLen(ctx, key).Result()
		if err != nil {
			t.Fatalf("HLEN failed: %v", err)
		}
		if n != 3 {
			t.Fatalf("HLEN = %d, want %d", n, 3)
		}
	})

	t.Run("7. 字段做数值累加", func(t *testing.T) {
		key := "article:1001"
		field := "stats:views"
		rdb.Del(ctx, key)

		if err := rdb.HSet(ctx, key, field, 100).Err(); err != nil {
			t.Fatalf("HSET numeric field failed: %v", err)
		}

		v1, err := rdb.HIncrBy(ctx, key, field, 1).Result()
		if err != nil {
			t.Fatalf("first HINCRBY failed: %v", err)
		}
		if v1 != 101 {
			t.Fatalf("first HINCRBY = %d, want %d", v1, 101)
		}

		v2, err := rdb.HIncrBy(ctx, key, field, 20).Result()
		if err != nil {
			t.Fatalf("second HINCRBY failed: %v", err)
		}
		if v2 != 121 {
			t.Fatalf("second HINCRBY = %d, want %d", v2, 121)
		}

		got, err := rdb.HGet(ctx, key, field).Result()
		if err != nil {
			t.Fatalf("HGET numeric field failed: %v", err)
		}
		if got != "121" {
			t.Fatalf("HGET numeric field = %q, want %q", got, "121")
		}
	})
}
