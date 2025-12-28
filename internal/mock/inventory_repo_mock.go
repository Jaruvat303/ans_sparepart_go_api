package mocks

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/inventory"
	"context"

	"github.com/stretchr/testify/mock"
)

type InventoryRepository struct {
	mock.Mock
}

func NewMockInventoryRepository() *InventoryRepository {
	return &InventoryRepository{}
}

func (i *InventoryRepository) GetByID(ctx context.Context, id uint) (*domain.Inventory, error) {
	args := i.Called(ctx, id)
	if inv, ok := args.Get(0).(*domain.Inventory); ok {
		return inv, args.Error(1)
	}
	return nil, args.Error(1)
}

func (i *InventoryRepository) GetByProductID(ctx context.Context, id uint) (*domain.Inventory, error) {
	args := i.Called(ctx, id)
	if inv, ok := args.Get(0).(*domain.Inventory); ok {
		return inv, args.Error(1)
	}
	return nil, args.Error(1)
}

func (i *InventoryRepository) List(ctx context.Context, q inventory.ListQuery) ([]*domain.Inventory, int64, error) {
	args := i.Called(ctx, q)

	var inv []*domain.Inventory
	if args.Get(0) != nil {
		inv = args.Get(0).([]*domain.Inventory)
	}
	count := args.Get(1).(int64)
	return inv, count, args.Error(2)
}

func (i *InventoryRepository) UpdateQuantity(ctx context.Context, id uint, quantity int) error {
	args := i.Called(ctx, id, quantity)
	return args.Error(0)
}

func (i *InventoryRepository) Create(ctx context.Context, inv *domain.Inventory) (*domain.Inventory, error) {
	args := i.Called(ctx, inv)
	if inv, ok := args.Get(0).(*domain.Inventory); ok {
		return inv, args.Error(1)
	}
	return nil, args.Error(1)
}

func (i *InventoryRepository) Delete(ctx context.Context, id uint) error {
	args := i.Called(ctx, id)
	return args.Error(0)
}
