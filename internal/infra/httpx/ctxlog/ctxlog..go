package ctxlog

import (
	"context"

	"go.uber.org/zap"
)

// ใช้ชนิด ctxKey ส่วนตัว ป่องกัน callision กับ ctxKey อื่นใน ctx
type ctxKey string

var loggerKey ctxKey = "ctx_logger"

// With ใส่ logger ลงใน context (ควรเรียกครั้งแรกใน middleware)
func With(ctx context.Context, zapLogger *zap.Logger) context.Context {
	if zapLogger == nil {
		zapLogger = zap.NewNop() // ป้องกันไม่ให้คืนค่า nil
	}

	return context.WithValue(ctx, loggerKey, zapLogger)
}

// From ดึง logger ที่ผูกกับ request ออกมาใช้จาก context
// ถ้าไม่มีให้คืน zap.NewNop() เพื่อหลีกเลี่ยง nil panic
func From(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok || l == nil {
		return zap.L() // fallback: global logger
	}
	return l
}

// WithFields สร้าง child logger ที่เพิ่มฟิลด์แล้วยัดกลับเข้า ctx
// ใช้ตอนอยากเพิ่ม context เฉพาะขั้น เช่นใส้ user_id หลัง auth
// func WithFields(ctx context.Context, field ...zap.Field) context.Context {
// 	return With(ctx, From(ctx).With(field...))
// }

// เพิ่ม fields ไปยัง logger ที่อยู่ใน Context
func AddFields(ctx context.Context, fields ...zap.Field) {
	l := From(ctx)
	newL := l.With(fields...)
	newCtx := With(ctx, newL)

	if setter, ok := ctx.(interface {
		SetUserContext(context.Context)
	}); ok {
		setter.SetUserContext(newCtx)
	}
}
