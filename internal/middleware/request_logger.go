package middleware

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func RequestLogger(root *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {

		start := time.Now()

		// request ID
		var reqID string
		if v, ok := c.Locals("requestid").(string); ok {
			reqID = v
		} else {
			// Fallback: generate requestID
			reqID = uuid.NewString()
			// เก็บที่ Locals
			c.Locals("requestid", reqID)
		}

		// สร้าง logger ใหม่สำหรับ request นี้
		reqLogger := root.With(
			zap.String("request_id", reqID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.String("ua", string(c.Request().Header.UserAgent())),
		)

		// Inject logger -> context
		ctx := ctxlog.With(c.UserContext(), reqLogger)
		c.SetUserContext(ctx)

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		log := ctxlog.From(c.UserContext())
		log.Info("http",
			zap.Int("status", status),
			zap.Duration("latency", latency),
		)

		return err
	}
}
