package product

import (
	"ans-spareparts-api/internal/features/category"
	"ans-spareparts-api/internal/features/inventory"
)

type CreateInput struct {
	Name        string
	Description string
	Price       float64
	SKU         string
	CategoryID  uint
}

type UpdateInput struct {
	Name        *string
	Description *string
	SKU         *string
	Price       *float64
	CategoryID  *uint
}

type ListQuery struct {
	Search string
	Limit  int
	Offset int
	Sort   string
}

type Item struct {
	ID          uint
	Name        string
	Description string
	SKU         string
	Price       float64
	IsActive    bool
	CategoryID  uint
	Category    category.CategoryResponse
	Inventory   inventory.InventoryResponse
}

type ItemLite struct {
	ID         uint
	Name       string
	SKU        string
	Price      float64
	CategoryID uint
	Active     bool
}

type ListOutput struct {
	Items []*ItemLite
	Total int64
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	SKU         string  `json:"sku"`
	Price       float64 `json:"price"`
	CategoryID  uint    `json:"category_id"`
}

type UpdateProductRequest struct {
	Name        *string
	Description *string
	SKU         *string
	Price       *float64
	CategoryID  *uint
}

type ProductDetailResponse struct {
	ID          uint
	Name        string
	Description string
	SKU         string
	Price       float64
	CategoryID  uint
	Category    category.CategoryResponse
	Inventory   inventory.InventoryResponse
}

type LiteProductResponse struct {
	ID         uint
	Name       string
	Price      float64
	SKU        string
	CategoryID uint
}

type ProductListResponse struct {
	Products []*LiteProductResponse
	Total    int64
}
