package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TeamMeng/go-demo/webook/internal/domain"
	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExist = redis.Nil

type UserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.Key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.User{}, ErrKeyNotExist
		}
		return domain.User{}, err
	}

	var user domain.User
	err = json.Unmarshal([]byte(val), &user)

	return user, nil
}

func (cache *UserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}

	key := cache.Key(u.Id)

	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UserCache) Key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
