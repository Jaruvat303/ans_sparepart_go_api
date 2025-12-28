package middleware

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Recover(root *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log := ctxlog.From(c.UserContext())
				log.Error("panic recoverd",
					zap.Any("recover", r),
				)
				err = fiber.ErrInternalServerError
			}
		}()

		return c.Next()
	}
}
