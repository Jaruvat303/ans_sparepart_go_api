package product

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/internal/infra/jwtx"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/response"
	"ans-spareparts-api/pkg/utils"
	"fmt"
	"strconv"

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

// CreateProduct godoc
// @Summary Create a new product
// @Descriton Create a new product (admin/mamager only)
// @Tag products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product creation request"
// @Success 201 {object} ProductDetailResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 409 {object} response.ErrorBody
// @Security BearerAuth
// @Router  /products [post]
func (h *Handler) CreateProduct(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	userClaims := c.Locals("user").(*jwtx.Claims)

	// Check permissions
	if userClaims.Role != "admin" && userClaims.Role != "manager" {
		log.Warn("handler.product.create.permission.not_allow",
			zap.Uint("user_id", userClaims.UserID),
			zap.String("role", userClaims.Role),
		)

		return response.Error(
			c, fiber.StatusForbidden, "FORBIDDEN", "insufficient permission",
		)
	}

	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn("handler.product.create.invalid_body", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid request body",
		)
	}

	product, err := h.service.CreateProduct(ctx, CreateInput{
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		if err == apperror.ErrConflict {
			return response.Error(
				c, fiber.StatusConflict, "CONFLICT", "product alreadt exist",
			)
		}

		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "category not found",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.Created(c, ProductDetailResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		SKU:         product.SKU,
		Price:       product.Price,
		CategoryID:  product.CategoryID,
		Category:    product.Category,
		Inventory:   product.Inventory,
	})
}

// GetProductDetail godoc
// @Summary Get product by ID
// @Description Get product details by product ID
// @Tag products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} ProductDetailResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Security BearerAuth
// @Router /products/{id} [get]
func (h *Handler) GetProductDetail(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	productID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		log.Warn("handler.product.getdetail.invalid_id", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid productID",
		)
	}

	product, err := h.service.GetProductDetail(ctx, uint(productID))
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "product not found",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.OK(c, ProductDetailResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		SKU:         product.SKU,
		Price:       product.Price,
		CategoryID:  product.CategoryID,
		Category:    product.Category,
		Inventory:   product.Inventory,
	})
}

// List godoc
// @Summary Get all products
// @Description Getall products with pagination
// @Tags products
// @Accept json
// @Product json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} ProductListResponse
// @Failure 500 {object} response.ErrorBody
// @Security BearerAuth
// @Router /products [get]
func (h *Handler) List(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)

	// init query
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		log.Warn("handler.product.list.invalid_input.limit", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid limit request",
		)
	}
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		log.Warn("handler.product.list.invalid_input.offset", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid offset request",
		)
	}
	sort := c.Query("sort", "ASC")
	search := c.Query("search", "")

	// Validate pagination parameters
	limit, offset = utils.NormalizePagination(limit, offset)

	products, err := h.service.List(ctx, ListQuery{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
		Search: search,
	})
	if err != nil {
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)

	}

	fmt.Printf("Product Output %v", products)
	var res = make([]*LiteProductResponse, len(products.Items))
	for i, r := range products.Items {
		res[i] = &LiteProductResponse{
			ID:         r.ID,
			Name:       r.Name,
			Price:      r.Price,
			CategoryID: r.CategoryID,
		}
	}

	// Return response with metadata
	return response.OK(c, ProductListResponse{Products: res, Total: products.Total})
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update product details (admin/manager only)
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body UpdateProductRequest true "Product update data"
// @Success 200 {object} ProductDetailResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 409 {object} response.ErrorBody
// @Security BearerAuth
// @Router /products/{id} [put]
func (h *Handler) UpdateProduct(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	userClaims := c.Locals("user").(*jwtx.Claims)

	// Check permissions
	if userClaims.Role != "admin" && userClaims.Role != "manager" {
		log.Warn("handler.product.update.permission,not_allow",
			zap.Uint("user_id", userClaims.UserID),
			zap.String("role", userClaims.Role),
		)
		return response.Error(c, fiber.StatusForbidden, "FORBIDDEN", "insufficient permission")
	}

	productID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		log.Warn("handler.product.update.invalid_id", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid productID",
		)
	}

	var req UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn("handler.product.update.invalid_body", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid body request",
		)
	}

	p, err := h.service.UpdateProduct(ctx, uint(productID), UpdateInput{
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "product not found",
			)
		}

		if err == apperror.ErrConflict {
			return response.Error(
				c, fiber.StatusConflict, "CONFLICT", "product sku already",
			)
		}

		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.OK(c, ProductDetailResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		SKU:         p.SKU,
		Price:       p.Price,
		CategoryID:  p.CategoryID,
		Category:    p.Category,
		Inventory:   p.Inventory,
	})

}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete a product (admin only)
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "ProductID"
// @Success 200 "No Content"
// @Failure 400 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Security BearerAuth
// @Router /products/{id} [delete]
func (h *Handler) DeleteProduct(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	userClaims := c.Locals("user").(*jwtx.Claims)

	// Check permissions
	if userClaims.Role != "admin" && userClaims.Role != "manager" {
		log.Warn("handler.product.update.permission,not_allow",
			zap.Uint("user_id", userClaims.UserID),
			zap.String("role", userClaims.Role),
		)
		return response.Error(c, fiber.StatusForbidden, "FORBIDDEN", "insufficient permission")
	}

	productID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		log.Warn("handler.product.delete.invalid_id", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid productID",
		)
	}

	if err := h.service.DeleteProduct(ctx, uint(productID)); err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "product not found",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.NoContent(c)
}
