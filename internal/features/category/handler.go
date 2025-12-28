package category

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/internal/infra/jwtx"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/response"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	categoryService Service
}

func NewHandler(categoryService Service) *Handler {
	return &Handler{categoryService: categoryService}
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new product category (admin/manager only)
// @Tags categories
// @Accept json
// @Produce json
// @Param category body Item true "Category creaion request"
// @Success 201 {object} CategoryResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Security BearerAuth
// @Router /categories [post]
func (h *Handler) CreateCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()  // ดึง ข้อมูลที่ผูกไว้กับ user Context
	log := ctxlog.From(ctx) // ดึง logger ที่ผูกกับ request id/user id
	userClaims := c.Locals("user").(*jwtx.Claims)

	// Check permissions Only admin or manager can create categories
	if userClaims.Role != "admin" && userClaims.Role != "manager" {
		log.Warn("handler.category.create.permission.not_allow",
			zap.Uint("user_id", userClaims.UserID),
			zap.String("role", userClaims.Role))
		return response.Error(
			c, fiber.StatusForbidden, "FORBIDDEN", "Insufficient permissions",
		)
	}

	var req CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Print("Invalid Request Body")
		log.Warn("handler.category.create.invalid_input", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid request body",
		)
	}

	category, err := h.categoryService.CreateCategory(ctx, req)
	if err != nil {
		if err == apperror.ErrConflict {
			return response.Error(
				c, fiber.StatusConflict, "CONFLICT", "category name already exist",
			)
		}

		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.Created(c, CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
	})
}

// GetCategory godoc
// @Summary Get Category by ID
// @Description Get category details by category ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} CategoryResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Security BearerAuth
// @Router /categories/{id} [get]
func (h *Handler) GetCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()  // ดึง ข้อมูลที่ผูกไว้กับ user Context
	log := ctxlog.From(ctx) // ดึง logger ที่ผูกกับ request id/user id
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		log.Warn("handler.category.get.invalid_input", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid category id",
		)
	}

	category, err := h.categoryService.GetCategoryByID(ctx, uint(categoryID))
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "category not found",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)

	}

	return response.OK(c, CategoryResponse{
		ID: category.ID,
		Name: category.Name,
	})
}

// List godoc
// @Summary Get all categories
// @Description Get all product categories
// Tags categories
// @Accept json
// @Product json
// @Success 200 {array} CategoryListResponse
// @Failure 500 {object} response.ErrorBody
// @Security BearerAuth
// @Router /categories [get]
func (h *Handler) List(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		log.Warn("handler.category.list.invalid_input.limit", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid limit request",
		)
	}
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		log.Warn("handler.category.list.invalid_input.offset", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid offset request",
		)
	}
	search := c.Query("search", "")

	out, err := h.categoryService.List(ctx, ListQuery{
		Limit:  limit,
		Offset: offset,
		Search: search,
	})
	if err != nil {
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	items := make([]*CategoryResponse, len(out.Items))
	for index, u := range out.Items {
		items[index] = &CategoryResponse{
			ID:   u.ID,
			Name: u.Name,
		}
	}

	return response.OK(c, CategoryListResponse{Categories: items, Total: out.Total})
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update category details (admin/manager only)
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body map[string]interface{} true "Category update data"
// @Success 200 {object} CategoryResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Security BearerAuth
// @Router /category/{id} [put]
func (h *Handler) UpdateCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	userClaims := c.Locals("user").(*jwtx.Claims)

	// Check Permission
	if userClaims.Role != "admin" && userClaims.Role != "manager" {
		log.Warn("handler.category.update.checkpermission.not_allow",
			zap.Uint("user_id", userClaims.UserID),
			zap.String("role", userClaims.Role))
		return response.Error(
			c, fiber.StatusForbidden, "FORBIDDEN", "insufficient permissions",
		)
	}

	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		log.Warn("handler.category.update.invalid_id", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid categoryID",
		)
	}

	var reqBody CategoryRequest
	if err := c.BodyParser(&reqBody); err != nil {
		log.Warn("handler.category.update.invalid_input", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invaid request body",
		)
	}

	category, err := h.categoryService.UpdateCategory(ctx, uint(categoryID), reqBody)
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "category not found",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.OK(c, CategoryResponse{
		ID: category.ID,
		Name: category.Name,
	})
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Delete a category (admin only)
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 "No Content"
// @Failure 400 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 409 {object} response.ErrorBody
// @Security BearerAuth
// @Router /categories/{id} [delete]
func (h *Handler) DeleteCategory(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	userClaims := c.Locals("user").(*jwtx.Claims)

	//Check permissions (only admin can delete categories)
	if userClaims.Role != "admin" {
		log.Warn("handler.category.delete.checkpermission.not_allow", zap.String("role", userClaims.Role))
		return response.Error(
			c, fiber.StatusForbidden, "FORBIDDEN", "Insufficient permissions",
		)
	}

	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		log.Warn("handler.category.delete.invalid_id", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid category id",
		)
	}

	err = h.categoryService.DeleteCategory(ctx, uint(categoryID))
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "category not found",
			)
		}

		if err == apperror.ErrConflict {
			return response.Error(
				c, fiber.StatusConflict, "CONFLICT", "category is in use",
			)
		}

		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.NoContent(c)
}
