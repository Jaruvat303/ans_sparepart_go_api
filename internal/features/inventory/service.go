package inventory

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/utils"

	"context"

	"go.uber.org/zap"
)

type Service interface {
	GetInventoryByID(ctx context.Context, inventoryID uint) (*Item, error)
	GetInventoryByProductID(ctx context.Context, productID uint) (*Item, error)
	List(ctx context.Context, q ListQuery) (*ListOutput, error)
	UpdateQuantity(ctx context.Context, id uint, input UpdateQuantityInput) (*Item, error)
}

type service struct {
	inventoryRepo Repository
}

func NewService(inventoryRepo Repository) Service {
	return &service{
		inventoryRepo: inventoryRepo,
	}
}

func (i *service) GetInventoryByID(ctx context.Context, ivnID uint) (*Item, error) {
	inventory, err := i.inventoryRepo.GetByID(ctx, ivnID)
	if err != nil {
		return nil, err
	}
	return &Item{ID: inventory.ID, ProductID: inventory.ProductID, Quantity: inventory.Quantity}, nil
}

func (i *service) GetInventoryByProductID(ctx context.Context, productID uint) (*Item, error) {
	inventory, err := i.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}
	return &Item{ID: inventory.ID, ProductID: inventory.ProductID, Quantity: inventory.Quantity}, nil
}

func (i *service) List(ctx context.Context, q ListQuery) (*ListOutput, error) {
	limit, offset := utils.NormalizePagination(q.Limit, q.Offset)
	rows, total, err := i.inventoryRepo.List(ctx, ListQuery{
		Limit:  limit,
		Offset: offset,
		Sort:   q.Sort,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*Item,  len(rows))
	for i, inv := range rows {
		items[i] = &Item{
			ID:        inv.ID,
			ProductID: inv.ProductID,
			Quantity:  inv.Quantity,
		}
	}
	return &ListOutput{Items: items, Total: total}, nil
}

func (i *service) UpdateQuantity(ctx context.Context, id uint, input UpdateQuantityInput) (*Item, error) {
	log := ctxlog.From(ctx)

	// check inventory exist
	inventory, err := i.inventoryRepo.GetByID(ctx, id)
	if inventory == nil {
		return nil, apperror.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	// check validquantity quantity สามารถติดลบได้
	// เงื่อนไขจำนวนที่ลดต้องไม่มากกว่า ค่า b เพราะหลังจากลดค่าไปแล้วจะติดลบ
	if input.Quantity < 0 && input.Quantity*-1 > inventory.Quantity {
		return nil, apperror.ErrInsufficientStock
	}

	// Update stock
	if err := i.inventoryRepo.UpdateQuantity(ctx, input.ProductID, input.Quantity); err != nil {
		return nil, err
	}

	log.Info("inventory.quantity.updated", zap.Uint("product_id", input.ProductID))
	return &Item{
		ID:        inventory.ID,
		ProductID: inventory.ProductID,
		Quantity:  inventory.Quantity + input.Quantity,
	}, nil
}
