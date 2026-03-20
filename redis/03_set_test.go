package redis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedisSet(t *testing.T) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis is not available: %v", err)
	}

	t.Run("1. 添加元素与去重", func(t *testing.T) {
		key := "user:1001:tags"
		rdb.Del(ctx, key)

		n, err := rdb.SAdd(ctx, key, "redis").Result()
		if err != nil {
			t.Fatalf("first SADD failed: %v", err)
		}
		if n != 1 {
			t.Fatalf("first SADD = %d, want %d", n, 1)
		}

		n, err = rdb.SAdd(ctx, key, "golang").Result()
		if err != nil {
			t.Fatalf("second SADD failed: %v", err)
		}
		if n != 1 {
			t.Fatalf("second SADD = %d, want %d", n, 1)
		}

		n, err = rdb.SAdd(ctx, key, "redis").Result()
		if err != nil {
			t.Fatalf("duplicate SADD failed: %v", err)
		}
		if n != 0 {
			t.Fatalf("duplicate SADD = %d, want %d", n, 0)
		}

		members, err := rdb.SMembers(ctx, key).Result()
		if err != nil {
			t.Fatalf("SMEMBERS failed: %v", err)
		}
		if len(members) != 2 {
			t.Fatalf("SMEMBERS len = %d, want %d", len(members), 2)
		}

		want := map[string]bool{
			"redis":  true,
			"golang": true,
		}
		for _, member := range members {
			if !want[member] {
				t.Fatalf("unexpected member %q", member)
			}
		}
	})

	t.Run("2. 判断元素是否存在", func(t *testing.T) {
		key := "user:1002:tags"
		rdb.Del(ctx, key)

		if err := rdb.SAdd(ctx, key, "redis", "docker").Err(); err != nil {
			t.Fatalf("SADD failed: %v", err)
		}

		ok, err := rdb.SIsMember(ctx, key, "redis").Result()
		if err != nil {
			t.Fatalf("SISMEMBER existing member failed: %v", err)
		}
		if !ok {
			t.Fatalf("SISMEMBER existing member = %v, want true", ok)
		}

		ok, err = rdb.SIsMember(ctx, key, "mysql").Result()
		if err != nil {
			t.Fatalf("SISMEMBER missing member failed: %v", err)
		}
		if ok {
			t.Fatalf("SISMEMBER missing member = %v, want false", ok)
		}
	})

	t.Run("3. 删除元素", func(t *testing.T) {
		key := "user:1003:tags"
		rdb.Del(ctx, key)

		if err := rdb.SAdd(ctx, key, "redis", "golang").Err(); err != nil {
			t.Fatalf("SADD failed: %v", err)
		}

		n, err := rdb.SRem(ctx, key, "golang").Result()
		if err != nil {
			t.Fatalf("existing SREM failed: %v", err)
		}
		if n != 1 {
			t.Fatalf("existing SREM = %d, want %d", n, 1)
		}

		ok, err := rdb.SIsMember(ctx, key, "golang").Result()
		if err != nil {
			t.Fatalf("SISMEMBER after SREM failed: %v", err)
		}
		if ok {
			t.Fatalf("member still exists after SREM")
		}

		n, err = rdb.SRem(ctx, key, "mysql").Result()
		if err != nil {
			t.Fatalf("missing SREM failed: %v", err)
		}
		if n != 0 {
			t.Fatalf("missing SREM = %d, want %d", n, 0)
		}
	})

	t.Run("4. 获取集合大小", func(t *testing.T) {
		key := "user:1004:tags"
		rdb.Del(ctx, key)

		if err := rdb.SAdd(ctx, key, "redis", "golang", "docker").Err(); err != nil {
			t.Fatalf("SADD failed: %v", err)
		}

		n, err := rdb.SCard(ctx, key).Result()
		if err != nil {
			t.Fatalf("SCARD failed: %v", err)
		}
		if n != 3 {
			t.Fatalf("SCARD = %d, want %d", n, 3)
		}
	})

	t.Run("5. 集合运算", func(t *testing.T) {
		key1 := "user:1005:skills"
		key2 := "user:1006:skills"
		rdb.Del(ctx, key1, key2)

		if err := rdb.SAdd(ctx, key1, "redis", "golang", "docker").Err(); err != nil {
			t.Fatalf("SADD key1 failed: %v", err)
		}
		if err := rdb.SAdd(ctx, key2, "redis", "mysql", "docker").Err(); err != nil {
			t.Fatalf("SADD key2 failed: %v", err)
		}

		intersection, err := rdb.SInter(ctx, key1, key2).Result()
		if err != nil {
			t.Fatalf("SINTER failed: %v", err)
		}
		if len(intersection) != 2 {
			t.Fatalf("SINTER len = %d, want %d", len(intersection), 2)
		}

		wantIntersection := map[string]bool{
			"redis":  true,
			"docker": true,
		}
		for _, member := range intersection {
			if !wantIntersection[member] {
				t.Fatalf("unexpected SINTER member %q", member)
			}
		}

		diff, err := rdb.SDiff(ctx, key1, key2).Result()
		if err != nil {
			t.Fatalf("SDIFF failed: %v", err)
		}
		if len(diff) != 1 || diff[0] != "golang" {
			t.Fatalf("SDIFF = %v, want [golang]", diff)
		}
	})
}
