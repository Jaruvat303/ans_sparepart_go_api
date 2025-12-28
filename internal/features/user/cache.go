package user

import (
	"ans-spareparts-api/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type cacheLayer struct {
	rdb *redis.Client
	ttl time.Duration
}

func newCache(rdb *redis.Client, ttl time.Duration) *cacheLayer {
	if rdb == nil || ttl <= 0 {
		return nil
	}

	return &cacheLayer{
		rdb: rdb,
		ttl: ttl,
	}
}

func (c *cacheLayer) keyByID(uID uint) string {
	return fmt.Sprintf("user:id:%d", uID)
}

func (c *cacheLayer) keyByUsername(username string) string {
	return fmt.Sprintf("user:id:%s", username)
}

func (c *cacheLayer) keyByEmail(email string) string {
	return fmt.Sprintf("user:id:%s", email)
}

func (c *cacheLayer) getByKey(ctx context.Context, key string) (*domain.User, bool, error) {
	if c == nil {
		return nil, false, nil
	}

	b, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false, err
	}

	var user domain.User
	if err := json.Unmarshal(b, &user); err != nil {
		return nil, false, err
	}

	return &user, true, nil
}

func (c *cacheLayer) set(ctx context.Context, key string, user *domain.User) error {
	if c == nil {
		return nil
	}

	b, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return c.rdb.Set(ctx, key, b, c.ttl).Err()
}

func (c *cacheLayer) del(ctx context.Context, key ...string) error {
	if c == nil || len(key) == 0 {
		return nil
	}

	return c.rdb.Del(ctx, key...).Err()
}
