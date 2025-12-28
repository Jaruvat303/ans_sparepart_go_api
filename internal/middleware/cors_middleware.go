package middleware

import (
	"ans-spareparts-api/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

// RegisterSecurity ติดตั้ง middleware ด้านความปลอดภัย
func RegisterSecurity(app *fiber.App, cfg *config.Config) {

	// helmet: ป้องกันหลายอย่างผ่าน Security Headers
	app.Use(helmet.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.HTTP.CORSAllowOrigins,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTION",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
		MaxAge:           3600, // 1 Hour
	}))

	// Rate Limiter: ป้องกัน brute force / flood attack
	// app.Use(limiter.New(limiter.Config{
	// 	Max:        100,             // 100 request
	// 	Expiration: 1 * time.Minute, // ภายใน 1 นาที

	// 	// response
	// 	LimitReached: func(c *fiber.Ctx) error {
	// 		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
	// 			"error": "limiter 100 request per 1 minues",
	// 		})
	// 	},
	// }))
}
