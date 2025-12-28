package main

// Swagger annotation
// @title ANS Sparepart API
// @version 1.0
// @description A Point of sale system API built with Clean Architechture
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email suppourt@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization Type "Bearer" followed by a space and JWT token.
import (
	"ans-spareparts-api/config"
	"ans-spareparts-api/internal/features/auth"
	"ans-spareparts-api/internal/features/category"
	"ans-spareparts-api/internal/features/inventory"
	"ans-spareparts-api/internal/features/product"
	"ans-spareparts-api/internal/features/user"
	"ans-spareparts-api/internal/infra/database"
	"ans-spareparts-api/internal/infra/hash"
	"ans-spareparts-api/internal/infra/jwtx"
	"ans-spareparts-api/internal/infra/logger"
	"ans-spareparts-api/internal/infra/redisx"
	"ans-spareparts-api/internal/middleware"
	"ans-spareparts-api/internal/router"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "ans-spareparts-api/docs"
)

func main() {
	cfg := config.Load()
	// logger
	rootLogger, atomicLvl, err := logger.New(logger.Options{
		Mode:             cfg.App.Mode,
		Level:            cfg.Log.Level,
		Service:          cfg.App.Name,
		Version:          cfg.App.Version,
		TimeLayout:       cfg.Log.TimeLayout,
		EnableSampling:   cfg.Log.EnableSampling,
		SampleInitial:    cfg.Log.SampleInitial,
		SampleThereafter: cfg.Log.SampleThereafter,
		AddCaller:        cfg.Log.AddCaller,
		StackOnError:     cfg.Log.StackOnError,
	})
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer rootLogger.Sync()

	// Connection Database (PostgresSQL)
	db := database.New(rootLogger, cfg)

	// Connection Redis
	rdb := redisx.New(redisx.Config{
		Addr:          cfg.Redis.Addr,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.DB,
		DialTimeout:   5 * time.Second,
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  3 * time.Second,
		PoolSize:      20,
		MinIdleConns:  2,
		SlowThreshold: 200 * time.Millisecond,
	}, rootLogger)

	// Initialize JWT manager พร้อม redis สำหรับ blacklist
	tokenManager := jwtx.NewJWTManager(jwtx.Options{
		Secret:   cfg.JWT.Secret,
		Issuer:   cfg.App.Name,
		Audience: "",                     // ถ้าไม่มี requirement ก็เว้นว่าง
		Expiry:   cfg.JWT.AccessTokenTTL, //24 h
		Leeway:   30 * time.Second,       // กัน clock skew
	}, rdb)

	// init hasher (bcrypt)
	hasher := hash.NewBcrypt(12)

	// Initialize repositories
	userRepo := user.NewRepository(db, rdb, 5*time.Minute)
	productRepo := product.NewRepository(db, rdb, 30*time.Minute)
	categoryRepo := category.NewRepository(db, rdb, 24*time.Hour)
	inventoryRepo := inventory.NewRepository(db, rdb, 10*time.Hour)

	// Initialze usecases
	authUseCase := auth.NewService(userRepo, tokenManager, hasher, "cashier")
	userUseCase := user.NewService(userRepo)
	productUseCase := product.NewService(productRepo, categoryRepo, inventoryRepo)
	categoryUseCase := category.NewService(categoryRepo)
	inventoryUseCase := inventory.NewService(inventoryRepo)

	// Create fiber app
	app := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		ReadTimeout:           cfg.HTTP.ReadTimeout,
		WriteTimeout:          cfg.HTTP.WriteTimeout,
		BodyLimit:             cfg.HTTP.BodyLimit,
		DisableStartupMessage: true,
	})

	// global middleware
	// requestid
	app.Use(requestid.New())
	// reciver
	app.Use(middleware.Recover(rootLogger))
	// request logger (ใช้ zap + ctxlog + request_id)
	app.Use(middleware.RequestLogger(rootLogger))
	// secutiry headers, CORS
	middleware.RegisterSecurity(app, cfg)

	// swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// log level management
	app.Post("/admin/log-level", func(c *fiber.Ctx) error {
		var body struct {
			Level string `json:"level"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}

		lvl := zapcore.InfoLevel
		if err := lvl.UnmarshalText([]byte(body.Level)); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid level"})
		}

		atomicLvl.SetLevel(lvl)

		return c.JSON(fiber.Map{"update_to": lvl.String()})

	})

	// Initialize router
	router.RegisterRoutes(app, router.Deps{
		AuthUC:       authUseCase,
		UserUC:       userUseCase,
		ProductUC:    productUseCase,
		CategoryUC:   categoryUseCase,
		InventoryUC:  inventoryUseCase,
		TokenManager: tokenManager,
	})

	// --- Start Server (Graceful Shutdown Pattern)---
	// ใช้ GoRotine รัน Server แยก
	// เพื่อไม่ให้ไป Block การทำงานของ บรรทัดถัดไป (ตัวรอรับ Signal)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
		rootLogger.Info("Server is starting on port", zap.String("port", addr))
		if err := app.Listen(addr); err != nil {
			rootLogger.Fatal("fiber listen error", zap.Error(err))
		}
	}()

	// สร้าง Channal เพื่อดักจับ OS Signal
	// SIGNT = กด Ctrl + C
	// SIGTERM = คำสั่งปิดจาก Docker/Kubernetes
	cn := make(chan os.Signal, 1)
	signal.Notify(cn, syscall.SIGINT, syscall.SIGTERM)
	<-cn

	rootLogger.Info("shutting down server...")

	// ปิด fiber
	if err := app.Shutdown(); err != nil {
		rootLogger.Error("fiber shutdown error", zap.Error(err))
	}

	// ปิด database
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	// ปิด redis
	if err := rdb.Close(); err != nil {
		rootLogger.Warn("redis close error", zap.Error(err))
	}

	rootLogger.Info("server existed gracefully")

}
