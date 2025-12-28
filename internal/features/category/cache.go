package category

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

func (c *cacheLayer) keyID(id uint) string {
	return fmt.Sprintf("category:id:%d", id)
}

func (c *cacheLayer) keyName(name string) string {
	return fmt.Sprintf("category:name:%s", name)
}

func (c *cacheLayer) getByKey(ctx context.Context, key string) (*domain.Category, bool, error) {
	if c == nil {
		return nil, false, nil
	}

	b, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var category domain.Category
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, false, err
	}

	return &category, true, nil
}

func (c *cacheLayer) set(ctx context.Context, key string, category *domain.Category) error {
	if c == nil {
		return nil
	}

	b, err := json.Marshal(category)
	if err != nil {
		return nil
	}

	return c.rdb.Set(ctx, key, b, c.ttl).Err()
}

func (c *cacheLayer) del(ctx context.Context, key ...string) error {
	if c == nil {
		return nil
	}
	return c.rdb.Del(ctx, key...).Err()
}
