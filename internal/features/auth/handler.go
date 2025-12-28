package auth

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} map[string]string "Example: {"message": "register successfull"}"
// @Failure 400 {object} response.ErrorBody "BAD_REQUEST: Invalid JSON or Input"
// @Failure 409 {object} response.ErrorBody "CONFLICT: Username or Email already exists"
// @Failure 500 {object} response.ErrorBody "INTERNAL_ERROR: Server error"
// @Router /auth/register [post]

// swagger:route POST /auth/register auth registerUser
// สมัครสมาชิกใหม่
// response:
//
//	201:
func (h *Handler) Register(c *fiber.Ctx) error {
	ctx := c.UserContext()  // ดึง ข้อมูลที่ผูกไว้กับ user Context
	log := ctxlog.From(ctx) // ดึง logger ที่ผูกกับ request id/user id

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn("handler.auth.register.invalid_input", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid json body",
		)
	}

	_, err := h.service.Register(ctx, RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err == apperror.ErrConflict {
			return response.Error(
				c, fiber.StatusConflict, "CONFLICT", "user already exists",
			)
		}
		if err == apperror.ErrInvalidInput {
			return response.Error(
				c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid input",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error occurred",
		)
	}

	return response.OK(c, fiber.Map{
		"message": "register successfull",
	})
}

// Login godoc
// @Summary User login
// @Description Authentication user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentails"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Router /auth/login [post]
func (h *Handler) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn("hadler.auth.login.invalid_input", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid request body",
		)
	}

	_, token, err := h.service.Login(ctx, LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		if err == apperror.ErrUnauthorized || err == apperror.ErrUserForbidden {
			return response.Error(
				c, fiber.StatusUnauthorized, "UNAUTHERIZED", "invalid credentials",
			)
		}

		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.OK(c, LoginResponse{
		Token: token,
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout user and blacklist token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} response.ErrorBody
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *Handler) Logout(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)

	token := c.Get("Authorization")
	if token == "" {
		log.Warn("handler.auth.logout.missingtoken", zap.String("token", token))
		return response.Error(
			c, fiber.StatusUnauthorized, "UNAUTHERIZED", "missing token",
		)
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.service.Logout(ctx, token)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to logout")
	}

	return response.OK(c, fiber.Map{
		"message": "logout success",
	})
}
