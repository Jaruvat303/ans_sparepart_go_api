package logger

// Setup Zap Logger configuration

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Options ปรับได้ทั้ง Dev/Prod โดยไม่ต้องแก้หลายที่
type Options struct {
	Mode             string // "dev"|"prod"
	Level            string // "debug"|"info"|"warn"|"error"
	Service          string
	Version          string
	TimeLayout       string // ว่าง = ISO8601
	EnableSampling   bool
	SampleInitial    int  // เริ่มจับคู่ 100 รายการแรก (Default 100)
	SampleThereafter int  // หลังจากนั้นทุกๆ N รายการ (Default 100)
	AddCaller        bool // เปิดใช้งานการบันทึกชื่อไฟล์/บรรทัดที่เรียก Log
	StackOnError     bool // บันทึก Stack Trace เมื่อเกิด Error หรือสูงกว่า
}

func New(o Options) (*zap.Logger, zap.AtomicLevel, error) {
	// -- level (atomic) --
	// กำหนด Log Level ขั้นต่ำที่จะถูกบันทึก ระดับที่ต่ำกว่านี้จะไม่ถูกบันทึก
	atomic := zap.NewAtomicLevelAt(parseLevel(o.Level))

	// -- encoder config --
	// ใช้ค่าตั้งต้นของ Production Config
	encCfg := zap.NewProductionEncoderConfig()
	// ต้ั้งค่าการเข้ารหัส Field Key
	encCfg.MessageKey = "msg"
	encCfg.LevelKey = "level"
	encCfg.TimeKey = "ts"
	encCfg.CallerKey = "caller"

	// ถ้าเป็น Console Encoder (Dev Mode) จะแสดงระดับ Log เป็นตัวพิมพ์เล็กพร้อมสีสัน
	// ถ้าเป็น JSON Encoder (Prod Mode) จะแสดงเป็นตัวพิมพ์เล็กเท่านั้น
	encCfg.EncodeLevel = zapcore.LowercaseColorLevelEncoder

	// การตั้งค่า EncodeTime (Custom Time Layout):
	if o.TimeLayout == "" {
		// "2025-10-29T08:47:00.123Z"
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// ใช้รูปแบบตามที่กำหนด
		layout := o.TimeLayout
		encCfg.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			pae.AppendString(t.Format(layout))
		}
	}

	// เลือก Encoder (Dev vs. Prod)
	var encoder zapcore.Encoder
	if isDev(o.Mode) {
		// console ช่วยให้อ่านง่ายตอน Dev
		encoder = zapcore.NewConsoleEncoder(encCfg)
	} else {
		// ใช้ (รูปแบบ JSON) เพื่อรองรับ Structured Logging ซึ่งจำเป็นสำหรับการวิเคราะห์ Log โดยเครื่องมือภายนอก
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	// -- write sink: stdout --
	// กำหนดให้ Log ทั้งหมดถูกเขียนออกไปยัง Standard Output (os.Stdout) และใช้ AddSync เพื่อรับประกันว่าการเขียน Log เป็นแบบ Thread-safe
	ws := zapcore.AddSync(os.Stdout)

	// -- core with or without sampling --
	var core zapcore.Core = zapcore.NewCore(encoder, ws, atomic)
	// การสุ่ม Log จะเกิดขึ้นก็ต่อเมื่อ o.EnableSampling เป็นจริง และ o.Mode ไม่ใช่ "dev"
	// (ไม่ควรสุ่ม Log ใน Dev Mode เพราะต้องการเห็น Log ทั้งหมด)
	if o.EnableSampling && !isDev(o.Mode) {
		initial := o.SampleInitial
		if initial <= 0 {
			initial = 100
		}
		after := o.SampleThereafter
		if after <= 0 {
			after = 100
		}
		core = zapcore.NewSamplerWithOptions(core, time.Second, initial, after)
	}

	// -- options --
	opts := []zap.Option{}
	if o.AddCaller {
		// หากเปิดใช้งาน จะเพิ่มฟิลด์ที่บอกว่า Log ถูกเรียกจากไฟล์และบรรทัดใด
		opts = append(opts, zap.AddCaller())
	}
	if o.StackOnError {
		// หากเปิดใช้งาน จะทำการบันทึก Stack Trace ทั้งหมดลงใน Log Entry เมื่อ Log นั้นมีระดับตั้งแต่ zap.ErrorLevel ขึ้นไป
		// (ERROR, DPANIC, PANIC, FATAL) ช่วยให้ Debug ปัญหาที่เกิดจากการ Crash ได้อย่างรวดเร็ว
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	// sticky fields
	fields := make([]zap.Field, 0, 2)
	// ... เพิ่ม service และ version เข้าไปใน fields
	if o.Service != "" {
		fields = append(fields, zap.String("service", o.Service))
	}
	if o.Version != "" {
		fields = append(fields, zap.String("version", o.Version))
	}
	if len(fields) > 0 {
		opts = append(opts, zap.Fields(fields...))
	}

	// สร้าง zaplogger
	logger := zap.New(core, opts...)
	return logger, atomic, nil
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func isDev(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "dev", "development", "local":
		return true
	default:
		return false
	}
}
