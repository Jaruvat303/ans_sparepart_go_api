package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// ใช้ env ในการ auto type parsing

type Config struct {
	App   AppConfig
	HTTP  HPPTConfig
	DB    DBConfig
	Redis RedisConfig
	JWT   JWTConfig
	Log   LogConfig
	GORM  GormConfig
}

type AppConfig struct {
	Name    string `env:"APP_NAME" envDefault:"inventory-service"`
	Version string `env:"APP_VERSION" envDefault:"1.0.0"`
	Mode    string `env:"APP_MODE" envDefault:"dev"`
}

// HTTP (Fiber)
type HPPTConfig struct {
	Host             string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port             int           `env:"HTTP_PORT" envDefault:"8080"`
	BodyLimit        int           `env:"HTTP_BODY_LIMIT" envDefault:"8388608"` // 8MB
	ReadTimeout      time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout     time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	CORSAllowOrigins string        `env:"HTTP_CORS_ALLOW_ORIGINS" envDefault:"*"`
}

// Database (Postgres)
type DBConfig struct {
	Host            string        `env:"DB_HOST" envDefault:"localhost"`
	Port            int           `env:"DB_PORT" envDefault:"5432"`
	Name            string        `env:"DB_NAME" envDefault:"mydatabase"`
	User            string        `env:"DB_USER" envDefault:"myuser"`
	Password        string        `env:"DB_PASSWORD" envDefault:"mypassword"`
	SSLMode         string        `env:"DB_SSLMODE" envDefault:"disable"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN" envDefault:"20"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"DB_MAX_LIFETIME" envDefault:"5m"`
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

type JWTConfig struct {
	Secret         string        `env:"JWT_SECRET" envDefault:"supersecret"`
	AccessTokenTTL time.Duration `env:"JWT_TTL" envDefault:"24h"`
	BlacklistTTL   time.Duration `env:"JWT_BLACKLIST_TTL" envDefault:"24h"`
}

type LogConfig struct {
	Level            string `env:"LOG_LEVEL" envDefault:"info"`
	EnableSampling   bool   `env:"LOG_SAMPLING" envDefault:"true"`
	AddCaller        bool   `env:"LOG_CALLER" envDefault:"true"`
	StackOnError     bool   `env:"LOG_STACK_ON_ERROR" envDefault:"true"`
	SampleInitial    int    `env:"LOG_SAMPLE_INITIAL" envDefault:"100"`
	SampleThereafter int    `env:"LOG_SAMPLE_THEREAFTER" envDefault:"100"`
	TimeLayout       string `env:"LOG_TIME_LAYOUT" envDefault:"2006-01-02T15:04:05Z07:00"`
}

type GormConfig struct {
	SlowThreshold         time.Duration `env:"GORM_SLOW_THRESHOLD" envDefault:"200ms"`
	LogLevel              string        `env:"GORM_LOG_LEVEL" envDefault:"warn"` // silent|error|worn|info
	IngnoreRecordNotFound bool          `env:"GORM_IGNORE_RECORD_NOT_FOUND" envDefault:"true"`
	DisableForeignKey     bool          `env:"GORM_DISBLE_FOREIGN_KEY" envDefault:"true"`
	PreparedStmt          bool          `env:"GORM_PREPARED_STMT" envDefault:"false"`
	NamingSingularTable   bool          `env:"GORM_NAMING_SINGULAR" envDefault:"false"`
}

// Load เรียกใช้ใน Main.go: ถ้าผิดพลาดให้ Panic
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file, using system env or defaults")
	}
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}
	return &cfg
}
