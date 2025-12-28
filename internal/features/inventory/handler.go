package inventory

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/internal/infra/jwtx"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/response"
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

// GetInventoryByID godoc
// @Summary Get inventory by ID
// @Description Get inventory detail by inventory id
// @Tag inventory
// @Accept json
// @Produce json
// @Param id path int true "inventory ID"
// @Success 200 {object} InventoryListResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Security BearerAuth
// @Router /inventory/{inventoryid} [get]
func (h *Handler) GetInventoryByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	log.Warn("handler.inventory.getbyId.invalid_id", zap.Error(err))
	if err != nil {
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid inventory id request",
		)
	}

	item, err := h.service.GetInventoryByID(ctx, uint(id))
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "inventory not found",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}

	return response.OK(c, InventoryResponse{
		ID:        item.ID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
	})
}

// GetInventoryByProductID godoc
// @Summary Get inventory by product ID
// @Description Get inventory details by product ID
// @Tags inventory
// @Accept json
// @Produce json
// @Param productId path int true "Product ID"
// @Success 200 {object} InventoryResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 404 {object} response.ErrorBody
// @Security BearerAuth
// @Router /product/{productID}/inventory [get]
func (h *Handler) GetInventoryByProductID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)

	productID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		log.Warn("handler.inventory.getbyproductid.invalid_id", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid productID request",
		)
	}

	inventory, err := h.service.GetInventoryByProductID(ctx, uint(productID))
	if err != nil {
		if err == apperror.ErrNotFound {
			return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "inventory not found")
		}
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured")

	}
	return response.OK(c, InventoryResponse{
		ID:        inventory.ID,
		ProductID: inventory.ProductID,
		Quantity:  inventory.Quantity,
	})
}

// UpdateQuantity godoc
// @Summary Update quantity inventory
// @Description Update inventory stock (admin/manager only)
// @Tags inventory
// @Accept json
// @Param inventory body UpdateQuantityRequest true "Inventory update request"
// @Success 200 {object} InventoryResponse
// @Failure 400 {object} response.ErrorBody
// @Failure 403 {object} response.ErrorBody
// @Security BearerAuth
// @Router /inventory [patch]
func (h *Handler) UpdateQuantity(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	userClaims := c.Locals("user").(*jwtx.Claims)

	// Check permissions
	if userClaims.Role != "admin" && userClaims.Role != "manager" {
		log.Warn("inventory.update_quantity.permission_check.not_allow",
			zap.Uint("user_id", userClaims.UserID),
			zap.String("user_role", userClaims.Role))
		return response.Error(
			c, fiber.StatusForbidden, "FORBIDDEN", "Insufficeint permission",
		)
	}

	invID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		log.Warn("handler.inventory.update_quantity.invalid_id", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid inventoryid",
		)
	}

	var req UpdateQuantityRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn("inventory.update_quantity.invalid_input", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid body request",
		)
	}

	inventory, err := h.service.UpdateQuantity(ctx, uint(invID), UpdateQuantityInput{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})

	if err != nil {
		if err == apperror.ErrInsufficientStock {
			return response.Error(
				c, fiber.StatusUnprocessableEntity, "UNPROCESSABLE", "Insufficient Stock",
			)
		}
		if err == apperror.ErrNotFound {
			return response.Error(
				c, fiber.StatusNotFound, "NOT_FOUND", "inventory for product not found",
			)
		}
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occuted",
		)
	}

	return response.OK(c, InventoryResponse{
		ID:        inventory.ID,
		ProductID: inventory.ProductID,
		Quantity:  inventory.Quantity,
	})
}

// List godoc
// @Summary Get inventories from query
// @Description Get all inventory records (admin/manager only)
// @Tags inventory
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} InventoryListResponse
// @Failure 403 {object} response.ErrorBody
// @Failure 500 {object} response.ErrorBody
// @Security BearerAuth
// @Router /inventory [get]
func (h *Handler) List(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := ctxlog.From(ctx)
	userClaims := c.Locals("user").(*jwtx.Claims)

	// Check permissions
	if userClaims.Role != "admin" && userClaims.Role != "manager" {
		log.Warn("handler.inventory.list.permission.not_allow",
			zap.Uint("user_id", userClaims.UserID),
			zap.String("role", userClaims.Role),
		)
		return response.Error(
			c, fiber.StatusForbidden, "FORBIDDEN", "insufficient permission",
		)
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		log.Warn("handler.inventory.list.invalid_input.limit", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid limit request",
		)
	}
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		log.Warn("handler.inventory.list.invalid_input.offset", zap.Error(err))
		return response.Error(
			c, fiber.StatusBadRequest, "BAD_REQUEST", "invalid offset request",
		)
	}
	sort := c.Query("sort", "ASC")

	inventories, err := h.service.List(ctx, ListQuery{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	})
	if err != nil {
		return response.Error(
			c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal server occured",
		)
	}
	res := make([]*InventoryResponse, len(inventories.Items))
	for i, inv := range inventories.Items {
		res[i] = &InventoryResponse{
			ID:        inv.ID,
			ProductID: inv.ProductID,
			Quantity:  inv.Quantity,
		}

	}

	return response.OK(
		c, InventoryListResponse{
			Inventories: res,
			Total:       inventories.Total,
		},
	)
}
