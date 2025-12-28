package redisx

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Config: ผูกกับ config.RedisConfig ได้เลย
type Config struct {
	Addr          string
	Password      string
	DB            int
	DialTimeout   time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	PoolSize      int
	MinIdleConns  int
	SlowThreshold time.Duration // กำหนดเวลาตำสั่งช้า
}

// สร้าง Client + health check + ติดตั้ง slow hook
func New(cfg Config, log *zap.Logger) *redis.Client {
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 3 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 3 * time.Second
	}
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 20
	}
	if cfg.MinIdleConns == 0 {
		cfg.MinIdleConns = 2
	}
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = 200 * time.Millisecond
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// ติดตั้ง hook จับคำสั่งช้า (warn)
	rdb.AddHook(NewSlowHook(log, cfg.SlowThreshold))

	//health check ตอน start
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("redis connection failed", zap.Error(err), zap.String("addr", cfg.Addr))

	}

	log.Info("redis connected", zap.String("addr", cfg.Addr))
	return rdb
}

// HealthCheck: ใช้กับ /healthz หรือ rediness probe
func HealthCheck(ctx context.Context, rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}
