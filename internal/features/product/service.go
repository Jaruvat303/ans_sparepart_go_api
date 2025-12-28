package product

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/category"
	"ans-spareparts-api/internal/features/inventory"
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"errors"
	"fmt"

	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/utils"
	"strings"

	"context"

	"go.uber.org/zap"
)

type Service interface {
	CreateProduct(ctx context.Context, in CreateInput) (*Item, error)
	GetProductDetail(ctx context.Context, productID uint) (*Item, error)
	UpdateProduct(ctx context.Context, productID uint, update UpdateInput) (*Item, error)
	DeleteProduct(ctx context.Context, productID uint) error
	List(ctx context.Context, q ListQuery) (*ListOutput, error)
}

type service struct {
	productRepo   Repository
	categoryRepo  category.Repository
	inventoryRepo inventory.Repository
}

func NewService(
	productRepo Repository,
	categoryRepo category.Repository,
	inventoryRepo inventory.Repository,
) Service {
	return &service{
		productRepo:   productRepo,
		categoryRepo:  categoryRepo,
		inventoryRepo: inventoryRepo,
	}
}

// --- Validators ---
func sanitizeCreate(in CreateInput) error {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.SKU) == "" {
		return apperror.ErrInvalidInput
	}
	if in.Price < 0 || in.CategoryID <= 0 {
		return apperror.ErrInvalidInput
	}

	return nil
}

func sanitizeUpdate(in UpdateInput) error {
	if in.Price != nil && *in.Price < 0 {
		return apperror.ErrInvalidInput
	}
	if in.CategoryID != nil && *in.CategoryID <= 0 {
		return apperror.ErrInvalidInput
	}
	return nil
}

// --- Mappers ---
func toItem(p *domain.Product, c *domain.Category, inv *domain.Inventory) *Item {
	out := &Item{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		IsActive:    p.IsActive,
	}
	if c != nil {
		out.Category.ID = c.ID
		out.Category.Name = c.Name
	}
	if inv != nil {
		out.Inventory.ID = inv.ID
		out.Inventory.ProductID = inv.ProductID
		out.Inventory.Quantity = inv.Quantity
	}

	return out
}

func (i *service) GetProductDetail(ctx context.Context, productID uint) (*Item, error) {

	product, err := i.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if i.categoryRepo == nil {
		fmt.Printf("categoreRepo is nil")
	}
	category, err := i.categoryRepo.GetByID(ctx, product.CategoryID)
	if err != nil {
		return nil, err
	}

	inventory, err := i.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return toItem(product, category, inventory), nil
}

func (i *service) CreateProduct(ctx context.Context, in CreateInput) (*Item, error) {
	log := ctxlog.From(ctx)

	// data validtor
	if err := sanitizeCreate(in); err != nil {
		return nil, err
	}

	// SKU validate and Normalized
	sku, err := utils.ValidateAndNormalizeSKU(in.SKU)
	if err != nil {
		return nil, err
	}

	// Check if SKU already exists
	existingProduct, err := i.productRepo.GetBySKU(ctx, in.SKU)
	if existingProduct != nil {
		return nil, apperror.ErrConflict
	}
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	// Verify category exists
	category, err := i.categoryRepo.GetByID(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}

	// Create Product
	product := &domain.Product{
		Name:        utils.SanitizeString(in.Name),
		Description: utils.SanitizeString(in.Description),
		Price:       in.Price,
		SKU:         sku,
		CategoryID:  in.CategoryID,
		IsActive:    true,
	}

	// Create Product
	if err := i.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	// create inventory with productid
	inventory := &domain.Inventory{
		ProductID: product.ID,
		Quantity:  0,
	}

	inv, err := i.inventoryRepo.Create(ctx, inventory)
	if err != nil {
		return nil, err
	}

	log.Info("product.created", zap.Uint("id", product.ID), zap.String("sku", product.SKU))

	return toItem(product, category, inv), nil
}

func (i *service) UpdateProduct(ctx context.Context, productID uint, in UpdateInput) (*Item, error) {
	log := ctxlog.From(ctx)

	product, err := i.productRepo.GetByID(ctx, productID)
	if product == nil {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// data validator
	if err := sanitizeUpdate(in); err != nil {
		return nil, err
	}

	// Apply updates
	if in.Name != nil && in.Name != &product.Name {
		product.Name = utils.SanitizeString(*in.Name)
	}
	if in.Description != nil && in.Description != &product.Description {
		product.Description = utils.SanitizeString(*in.Description)
	}
	if in.SKU != nil && in.Description != &product.SKU {
		// SKU validate and Normalized
		sku, err := utils.ValidateAndNormalizeSKU(*in.SKU)
		if err != nil {
			return nil, err
		}

		// Check if SKU already exists
		existingProduct, err := i.productRepo.GetBySKU(ctx, *in.SKU)
		if existingProduct != nil {
			return nil, apperror.ErrConflict
		}
		if err != nil && !errors.Is(err, apperror.ErrNotFound) {
			return nil, err
		}

		product.SKU = sku
	}
	if in.Price != nil && in.Price != &product.Price {
		product.Price = *in.Price
	}
	if in.CategoryID != nil && in.CategoryID != &product.CategoryID {
		// verify category exists
		category, err := i.categoryRepo.GetByID(ctx, *in.CategoryID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, apperror.ErrNotFound
		}

		product.CategoryID = *in.CategoryID
		product.Category = *category
	}

	if err := i.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	inventory, err := i.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	log.Info("product.updated", zap.Uint("id", productID))
	return toItem(product, &product.Category, inventory), nil
}

func (i *service) DeleteProduct(ctx context.Context, productID uint) error {
	log := ctxlog.From(ctx)

	_, err := i.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// Delete product
	if err := i.productRepo.Delete(ctx, productID); err != nil {
		return err
	}

	// Delete inventory
	if err := i.inventoryRepo.Delete(ctx, productID); err != nil {
		return err
	}

	log.Info("product.Deledted", zap.Uint("productID", productID))
	return nil
}

func (i *service) List(ctx context.Context, q ListQuery) (*ListOutput, error) {
	// map query
	query := ListQuery{
		Search: q.Search,
		Limit:  q.Limit,
		Offset: q.Offset,
		Sort:   q.Sort,
	}
	products, total, err := i.productRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	items := make([]*ItemLite, 0, len(products))
	for _, p := range products {
		items = append(items, &ItemLite{
			ID:     p.ID,
			Name:   p.Name,
			SKU:    p.SKU,
			Price:  p.Price,
			Active: p.IsActive,
		})
	}

	return &ListOutput{Items: items, Total: total}, nil
}
