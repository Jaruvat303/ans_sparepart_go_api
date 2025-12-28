package cache

// Setup Redis connection

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisClient(config Config, zapLogger *zap.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20, // จำนวน Connection pool
		MinIdleConns: 2,  // เตรียมพร้อม 2 connection พร้อมใช้
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		zapLogger.Fatal("redis connection failed", zap.Error(err))
	}

	zapLogger.Info("redis connected", zap.String("addr", config.Addr))
	return rdb
}

// HealthCheck ใช้ตรวจสอบสถานะในระบบ Monitoring
// ใช้ใน /healthz หรือระบบ monitoring เช่น Kubernetes liveness probe
// ถ้า Redis down จะได้ตรวจจับและ restart pod
func HealthCheck(ctx context.Context, rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}
