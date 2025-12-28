package database

// Setup GORM connecttion

import (
	"ans-spareparts-api/config"
	"ans-spareparts-api/internal/infra/gormzap"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func New(z *zap.Logger, cfg *config.Config) *gorm.DB {
	// setup GORM Logger
	gz := gormzap.New(z, logger.Config{
		SlowThreshold:             cfg.GORM.SlowThreshold,
		LogLevel:                  gormzap.ParseLogLevel(cfg.GORM.LogLevel),
		IgnoreRecordNotFoundError: cfg.GORM.IngnoreRecordNotFound,
		Colorful:                  cfg.App.Mode == "dev",
	})

	// ORM Config
	ns := schema.NamingStrategy{}
	if cfg.GORM.NamingSingularTable {
		ns.SingularTable = true
	}

	gcfg := &gorm.Config{
		Logger:                                   gz,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: cfg.GORM.DisableForeignKey,
		NamingStrategy:                           ns,
		PrepareStmt:                              cfg.GORM.PreparedStmt,
	}

	// Dialector
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.User, cfg.DB.Password, cfg.DB.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), gcfg)
	if err != nil {
		z.Fatal("failed to open database", zap.Error(err))
	}

	// Connection Pool
	sqlDB, err := db.DB()
	if err != nil {
		z.Fatal("failed to unwrap sql.DB", zap.Error(err))
	}
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	z.Info("database connected",
		zap.String("host", cfg.DB.Host),
		zap.Int("port", cfg.DB.Port),
		zap.String("db", cfg.DB.Name),
	)

	return db
}
