package product

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
	return &cacheLayer{rdb: rdb, ttl: ttl}
}

func (c *cacheLayer) keyByID(id uint) string {
	return fmt.Sprintf("product:id:%d", id)
}

func (c *cacheLayer) keyBySKU(sku string) string {
	return fmt.Sprintf("product:sku:%s", sku)
}

func (c *cacheLayer) getByKey(ctx context.Context, key string) (*domain.Product, bool, error) {
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
	var p domain.Product
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, false, err
	}
	return &p, true, nil
}

func (c *cacheLayer) set(ctx context.Context, key string, p *domain.Product) error {
	if c == nil {
		return nil
	}
	b, err := json.Marshal(p)
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
