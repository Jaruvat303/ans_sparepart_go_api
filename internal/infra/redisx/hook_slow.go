package redisx

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// slowHook: log คำสั่งที่ช้ากว่า threshold
type slowHook struct {
	log       *zap.Logger
	threshold time.Duration
}

func NewSlowHook(log *zap.Logger, threshold time.Duration) redis.Hook {
	if log == nil {
		log = zap.NewNop()
	}
	return &slowHook{log: log, threshold: threshold}
}

func (h *slowHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (h *slowHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	start, _ := ctx.Value(startKey{}).(time.Time)
	elapsed := time.Since(start)
	if elapsed > h.threshold {
		h.log.Warn("redis.slow",
			zap.String("cmd", cmd.Name()),
			zap.Duration("duration", elapsed),
			zap.String("args", compactArgs(cmd.Args())),
		)
	}
	if err := cmd.Err(); err != nil && err != redis.Nil {
		h.log.Error("redis.error", zap.String("cmd", cmd.Name()), zap.Error(err))
	}
	return nil
}

// Pipeline hook (รองรับ MULTI/PIPELINE)
func (h *slowHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (h *slowHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	start, _ := ctx.Value(startKey{}).(time.Time)
	elapsed := time.Since(start)
	if elapsed > h.threshold {
		h.log.Warn("redis.pipeline.slow",
			zap.Int("n", len(cmds)),
			zap.Duration("duration", elapsed),
		)
	}

	// log error ของแต่ละ cmd
	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil && err != redis.Nil {
			h.log.Error("redis.pipeline.error", zap.String("cmd", cmd.Name()), zap.Error(err))
		}
	}

	return nil
}

type startKey struct{}

func compactArgs(args []interface{}) string {
	// ย่อ args เพื่อไม่ให้ log ่ยาว/รัว secret; ใส่แค่ชนิดและจำนวน
	// เช่น "key=product:id:10, extra=2"
	if len(args) == 0 {
		return ""
	}
	key := ""
	if s, ok := args[1].(string); ok {
		key = s
	}
	return "key=" + key + ", extra=" + itoa(len(args)-2)
}

// tiny itoa (ไม่ต้อง import strconv สำหรับแค่แปลงเล็ก ๆ)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	d := [20]byte{}
	i := len(d)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		d[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		d[i] = '-'
	}
	return string(d[i:])
}
