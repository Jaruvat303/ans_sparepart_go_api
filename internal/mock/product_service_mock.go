package mocks

import (
	"ans-spareparts-api/internal/features/product"
	"context"

	"github.com/stretchr/testify/mock"
)

type ProductService struct {
	mock.Mock
}

func NewProductService() *ProductService {
	return &ProductService{}
}

// type Service interface {
// 	CreateProduct(ctx context.Context, in CreateInput) (*Item, error)
// 	GetProductDetail(ctx context.Context, productID uint) (*Item, error)
// 	UpdateProduct(ctx context.Context, productID uint, update UpdateInput) (*Item, error)
// 	DeleteProduct(ctx context.Context, productID uint) error
// 	List(ctx context.Context, q ListQuery) (*ListOutput, error)
// }

func (m *ProductService) CreateProduct(ctx context.Context, input product.CreateInput) (*product.Item, error) {
	args := m.Called(ctx, input)
	if value, ok := args.Get(0).(*product.Item); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ProductService) GetProductDetail(ctx context.Context, productID uint) (*product.Item, error) {
	args := m.Called(ctx, productID)
	if value, ok := args.Get(0).(*product.Item); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ProductService) UpdateProduct(ctx context.Context, productID uint, update product.UpdateInput) (*product.Item, error) {
	args := m.Called(ctx, productID, update)
	if value, ok := args.Get(0).(*product.Item); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ProductService) DeleteProduct(ctx context.Context, productID uint) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *ProductService) List(ctx context.Context, query product.ListQuery) (*product.ListOutput, error) {
	args := m.Called(ctx, query)
	if value, ok := args.Get(0).(*product.ListOutput); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}
