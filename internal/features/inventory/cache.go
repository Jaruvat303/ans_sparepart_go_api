package inventory

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

func (c *cacheLayer) keyProductID(id uint) string {
	return fmt.Sprintf("inventory:product_id:%d", id)
}

func (c *cacheLayer) keyID(id uint) string {
	return fmt.Sprintf("inventory:product_id:%d", id)
}

func (c *cacheLayer) getByKey(ctx context.Context, key string) (*domain.Inventory, bool, error) {
	if c == nil {
		return nil, false, nil
	}

	b, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false, err
	}

	var inventory domain.Inventory
	if err := json.Unmarshal(b, &inventory); err != nil {
		return nil, false, err
	}

	return &inventory, true, nil

}

func (c *cacheLayer) set(ctx context.Context, key string, inventory *domain.Inventory) error {
	if c == nil {
		return nil
	}

	b, err := json.Marshal(inventory)
	if err != nil {
		return err
	}

	return c.rdb.Set(ctx, key, b, c.ttl).Err()
}

func (c *cacheLayer) del(ctx context.Context, key string) error {
	if c == nil {
		return nil
	}

	return c.rdb.Del(ctx, key).Err()
}
