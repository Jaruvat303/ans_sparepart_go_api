package middleware

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/internal/infra/jwtx"
	"ans-spareparts-api/pkg/response"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RequireAuth
// ตรวจสอบ Bearer token, validate JWT, ตรวจสอบ blacklist, inject claims ลง ctx
func RequireAuth(tm jwtx.TokenManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := jwtx.ExtractBearer(c.Get("Authorization"))
		if token == "" {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token")
		}
		claims, err := tm.ValidateToken(token)
		if err != nil {
			switch err {
			case jwtx.ErrExpiredToken:
				return response.Error(c, fiber.StatusUnauthorized, "TOKEN_EXPIRED", "token expired")
			default:
				return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
			}
		}

		// blacklist?
		if revoked, err := tm.IsBlacklisted(c.UserContext(), claims.ID); err == nil && revoked {
			return response.Error(c, fiber.StatusUnauthorized, "UNTHORIZED", "token revoked")
		}

		ctx := jwtx.InjectClaims(c.UserContext(), claims)
		c.SetUserContext(ctx)

		// ผูกข้อมูลลง context/log
		ctxlog.AddFields(ctx,
			zap.Uint("auth.user_id", claims.UserID),
			zap.String("auth.username", claims.Username),
			zap.String("auth.role", claims.Role),
		)

		return c.Next()
	}
}

// RequireRole checks if the authenticated user has at least one of the given roles.
func RequireRole(roles ...string) fiber.Handler {
	// เตรียม map เพื่อ lookup role
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[strings.ToLower(r)] = true
	}

	return func(c *fiber.Ctx) error {
		log := ctxlog.From(c.UserContext())

		claims, ok := jwtx.FormContext(c.UserContext())
		if !ok {
			// กรณีลืมใส่ RequireAuth ก่อน RequireRole
			log.Error("middleware.require_role.no_claims")
			return response.Error(
				c, fiber.StatusUnauthorized, "UNAUTHRIZED", "missing authentication",
			)
		}

		userRole := strings.ToLower(claims.Role)

		if !allowed[userRole] {
			log.Warn("middleware.require_role.forbidden",
				zap.String("require_roles", strings.Join(roles, ",")),
				zap.String("user_role", claims.Role),
			)

			return response.Error(
				c, fiber.StatusForbidden, "FORBIDDEN", "insufficient role",
			)
		}
		return c.Next()
	}
}
