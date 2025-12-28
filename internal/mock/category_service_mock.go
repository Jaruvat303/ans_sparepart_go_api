package mocks

import (
	"ans-spareparts-api/internal/features/category"
	"context"

	"github.com/stretchr/testify/mock"
)

type CategoryService struct {
	mock.Mock
}

func NewMockCategoryService() *CategoryService {
	return &CategoryService{}
}

func (m *CategoryService) CreateCategory(ctx context.Context, req category.CategoryRequest) (*category.Item, error) {
	args := m.Called(ctx, req)
	if value, ok := args.Get(0).(*category.Item); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *CategoryService) GetCategoryByID(ctx context.Context, id uint) (*category.Item, error) {
	args := m.Called(ctx, id)
	if item, ok := args.Get(0).(*category.Item); ok {
		return item, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *CategoryService) GetCategoryByName(ctx context.Context, req category.CategoryRequest) (*category.Item, error) {
	args := m.Called(ctx, req)
	if item, ok := args.Get(0).(*category.Item); ok {
		return item, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *CategoryService) UpdateCategory(ctx context.Context, id uint, req category.CategoryRequest) (*category.Item, error) {
	args := m.Called(ctx, id, req)
	if item, ok := args.Get(0).(*category.Item); ok {
		return item, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *CategoryService) DeleteCategory(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *CategoryService) List(ctx context.Context, query category.ListQuery) (*category.ListOutput, error) {
	args := m.Called(ctx, query)
	if list, ok := args.Get(0).(*category.ListOutput); ok {
		return list, args.Error(1)
	}
	return nil, args.Error(1)
}
