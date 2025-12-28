package mocks

import (
	"ans-spareparts-api/internal/features/inventory"
	"context"

	"github.com/stretchr/testify/mock"
)

type InventoryService struct {
	mock.Mock
}

func NewInventoryService() *InventoryService {
	return &InventoryService{}
}

func (m *InventoryService) GetInventoryByID(ctx context.Context, invID uint) (*inventory.Item, error) {
	args := m.Called(ctx, invID)
	if value, ok := args.Get(0).(*inventory.Item); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *InventoryService) GetInventoryByProductID(ctx context.Context, pID uint) (*inventory.Item, error) {
	args := m.Called(ctx, pID)
	if value, ok := args.Get(0).(*inventory.Item); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *InventoryService) UpdateQuantity(ctx context.Context, ivnID uint, input inventory.UpdateQuantityInput) (*inventory.Item, error) {
	args := m.Called(ctx, ivnID, input)
	if value, ok := args.Get(0).(*inventory.Item); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *InventoryService) List(ctx context.Context, q inventory.ListQuery) (*inventory.ListOutput, error) {
	args := m.Called(ctx, q)
	if value, ok := args.Get(0).(*inventory.ListOutput); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}
