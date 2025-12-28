package mocks

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/product"
	"context"

	"github.com/stretchr/testify/mock"
)

type ProductRepository struct {
	mock.Mock
}

func NewMockProductRepository() *ProductRepository {
	return &ProductRepository{}
}

func (r *ProductRepository) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	args := r.Called(ctx, id)
	if product, ok := args.Get(0).(*domain.Product); ok {
		return product, args.Error(1)
	}
	return nil, args.Error(1)
}

func (r *ProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	args := r.Called(ctx, sku)

	if product, ok := args.Get(0).(*domain.Product); ok {
		return product, args.Error(1)
	}
	return nil, args.Error(1)
}

func (r *ProductRepository) List(ctx context.Context, query product.ListQuery) ([]*domain.Product, int64, error) {
	args := r.Called(ctx, query)

	var products []*domain.Product
	if args.Get(0) != nil {
		products = args.Get(0).([]*domain.Product)
	}

	totals := args.Get(1).(int64)
	return products, totals, args.Error(2)
}

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	args := r.Called(ctx, product)
	return args.Error(0)
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	args := r.Called(ctx, product)
	return args.Error(0)
}

func (r *ProductRepository) Delete(ctx context.Context, id uint) error {
	args := r.Called(ctx, id)
	return args.Error(0)
}
