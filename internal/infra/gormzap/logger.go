package gormzap

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// สร้าง struct ที่ Implement interface หลักของ GORM Logger เพื่อให้ GORM สามารถใช้ได้
type Logger struct {
	zap    *zap.Logger
	config gormlogger.Config
}

func New(z *zap.Logger, cfg gormlogger.Config) *Logger {
	return &Logger{zap: z, config: cfg}
}
func (l *Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	n := *l
	n.config.LogLevel = level
	return &n
}

// การแมป Log ทั่วไป (Info, Warn, Error)เป็นรูปแบบ Json
// Logger จะใช้ Zap Sugared Logger ผ่าน .Sugar()

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel < gormlogger.Info {
		return
	}
	l.zap.Sugar().Infow("gorm.info", "msg", fmt.Sprintf(msg, data...))
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel < gormlogger.Warn {
		return
	}
	l.zap.Sugar().Warnw("gorm.warn", "msg", fmt.Sprintf(msg, data...))
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel < gormlogger.Error {
		return
	}
	l.zap.Sugar().Errorw("gorm.error", "msg", fmt.Sprintf(msg, data...))
}

// เมธอด Trace คือหัวใจหลักของ GORM Logger ที่ถูกเรียกทุกครั้งที่มีการ Execute คำสั่ง SQL
func (l *Logger) Trace(ctx context.Context, begin time.Time, fnc func() (string, int64), err error) {
	if l.config.LogLevel == gormlogger.Silent {
		return
	}
	// คำนวณ ระยะเวลาที่ใช้ไป ในการรัน Query
	elapsed := time.Since(begin)
	// เรียก Callback Function เพื่อดึง SQL Query String และ Row Count จำนวนแถวที่ได้รับผลกระทบ
	sql, rows := fnc()

	// สร้าง Logger Instance ใหม่ที่ มี Context โดยการแนบ Field สำคัญ 3 ตัวเข้าไปใน Log Line นี้
	// base := l.zap.With(
	// 	zap.Duration("duration", elapsed),
	// 	zap.Int64("row", rows),
	// )

	// การตัดสินใจ Log (Switch Case)
	switch {
	case err != nil && l.config.LogLevel >= gormlogger.Error:
		l.zap.Error("gorm.query.error",
			zap.Error(err),
			zap.Duration("duration", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.config.SlowThreshold != 0 && elapsed > l.config.SlowThreshold && l.config.LogLevel >= gormlogger.Warn:
		l.zap.Warn("gorm.query.slow",
			zap.Duration("duration", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.config.LogLevel == gormlogger.Info:
		l.zap.Debug("gorm.query",
			zap.Duration("duration", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}
