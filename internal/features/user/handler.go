package user

import (
	"ans-spareparts-api/internal/features/auth"
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/internal/infra/jwtx"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	userService Service
	authService auth.Service
}

func NewHandler(userService Service, authService auth.Service) *Handler {
	return &Handler{
		userService: userService,
		authService: authService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current authenticated user's profile
// @Tags User Profile
// @Accept json
// @Produce json
// @Security BeaterAuth
// @Produce json
// @Success 200 {Object} UserResponse
// @Failure 401 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Router /users/profile [get]

func (h *Handler) GetProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userClaims := c.Locals("user").(*jwtx.Claims)

	user, err := h.userService.GetUserProfile(ctx, userClaims.UserID)
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "user not found",
			)
		}

		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.OK(c, UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	})
}

// DeleteUser godoc
// @Summary Delete user account
// @Description Delete current authenticated user's profile
// @Tags User Profile
// @Accept json
// @Produce json
// @Success 200 "No Content"
// @Failure 400 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Security BearerAuth
// @Router /users/profile [delete]
func (h *Handler) DeleteProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	userClaims := c.Locals("user").(*jwtx.Claims)

	// Logout after delete user
	token := c.Get("Authorization")
	if token == "" {
		log.Warn("handler.user.delete.token.missing", zap.Uint("user_id", userClaims.UserID))
		return response.Error(
			c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing token",
		)
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.userService.DeleteUser(ctx, userClaims.UserID)
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "user not found",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	err = h.authService.Logout(ctx, token)
	if err != nil {
		if err == apperror.ErrInvalidToken {
			return response.Error(
				c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid token",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)

	}
	return response.NoContent(c)
}
