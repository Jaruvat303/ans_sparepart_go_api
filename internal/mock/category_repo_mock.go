package mocks

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/category"
	"context"

	"github.com/stretchr/testify/mock"
)

type CategoryRepository struct {
	mock.Mock
}

func NewMockCategoryRepository() *CategoryRepository {
	return &CategoryRepository{}
}

func (m *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	agrs := m.Called(ctx, category)
	return agrs.Error(0)
}

func (m *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *CategoryRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *CategoryRepository) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if cat, ok := args.Get(0).(*domain.Category); ok {
		return cat, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *CategoryRepository) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	args := m.Called(ctx, name)
	if cat, ok := args.Get(0).(*domain.Category); ok {
		return cat, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *CategoryRepository) List(ctx context.Context, q category.ListQuery) ([]*domain.Category, int64, error) {
	args := m.Called(ctx, q)

	var categories []*domain.Category
	if args.Get(0) != nil {
		categories = args.Get(0).([]*domain.Category)
	}
	count := args.Get(1).(int64)
	return categories, count, args.Error(2)
}
