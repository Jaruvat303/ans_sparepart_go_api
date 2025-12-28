package inventory

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/pkg/apperror"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	GetByID(ctx context.Context, invID uint) (*domain.Inventory, error)
	GetByProductID(ctx context.Context, productID uint) (*domain.Inventory, error)
	List(ctx context.Context, q ListQuery) ([]*domain.Inventory, int64, error)
	UpdateQuantity(ctx context.Context, productID uint, quantity int) error

	// ใช้ที่ Product Interactor เมื่อสร้าง Product หรือ ลบ Products
	Create(ctx context.Context, inventory *domain.Inventory) (*domain.Inventory, error)
	Delete(ctx context.Context, pId uint) error
}

type repository struct {
	db    *gorm.DB
	cache *cacheLayer
}

func NewRepository(db *gorm.DB, rdb *redis.Client, cacheExp time.Duration) Repository {
	return &repository{
		db:    db,
		cache: newCache(rdb, cacheExp),
	}
}

func (r *repository) GetByID(ctx context.Context, invID uint) (*domain.Inventory, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// cache check
	if cache, ok, err := r.cache.getByKey(ctx, r.cache.keyID(invID)); err == nil && ok {
		log.Debug("repo.inventory.getbyid_cache_hit", zap.Uint("inventory_id", cache.ID), zap.Duration("duration", time.Since(start)))
		return cache, nil
	} else if err != nil {
		log.Warn("repo.inventory.getbyid.cache_err", zap.Uint("inventory_id", invID), zap.Error(err))
	}

	// DB
	var inventory domain.Inventory
	if err := r.db.WithContext(ctx).First(&inventory, invID).Error; err != nil {
		m := apperror.MapDBError("repo.inventory.getbyid", err)
		log.Debug("repo.inventory.getbyid.db_err", zap.Uint("inventory_id", invID), zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyID(invID), &inventory); err != nil {
		log.Warn("repo.inventory.getByID.set_cache.fail", zap.Error(err))
	}

	log.Debug("repo.inventory.getByProductID.ok", zap.Uint("inventory_id", invID), zap.Duration("duration", time.Since(start)))
	return &inventory, nil

}

func (r *repository) GetByProductID(ctx context.Context, pID uint) (*domain.Inventory, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// cache check
	if cache, ok, err := r.cache.getByKey(ctx, r.cache.keyProductID(pID)); err == nil && ok {
		log.Debug("repo.inventory.getByProductID.cache_hit", zap.Uint("product_id", pID), zap.Duration("duration", time.Since(start)))
		return cache, nil
	} else if err != nil {
		log.Warn("repo.inventory.geyByProductID.cache_error")
	}

	// DB
	var inventory domain.Inventory
	if err := r.db.WithContext(ctx).First(&inventory, "product_id = ?", pID).Error; err != nil {
		m := apperror.MapDBError("repo.inventory.getByproductID", err)
		log.Debug("repo.inventory.getByProductID.db_fail", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyID(pID), &inventory); err != nil {
		log.Warn("repo.inventory.getByProductID.set_cache.fail", zap.Error(err))
	}

	log.Debug("repo.inventory.getByProductID.ok", zap.Uint("product_id", pID), zap.Duration("duration", time.Since(start)))
	return &inventory, nil
}

func (r *repository) List(ctx context.Context, q ListQuery) ([]*domain.Inventory, int64, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// Query Builder Session
	tx := r.db.WithContext(ctx).Find(&domain.Inventory{})

	// Count รวท
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		m := apperror.MapDBError("repo.inventory.list", err)
		log.Warn("repo.inventory.list.count_err", zap.Error(err))
		return nil, 0, m
	}

	if q.Sort != "" {
		tx = tx.Order(q.Sort)
	} else {
		tx = tx.Order("created_at DESC")
	}
	if q.Offset > 0 {
		tx = tx.Offset(q.Offset)
	}
	if q.Limit > 0 {
		tx = tx.Limit(q.Limit)
	}

	// Find with Query Builder Session
	var rows []*domain.Inventory
	if err := tx.Find(&rows).Error; err != nil {
		m := apperror.MapDBError("repo.inventory.list", err)
		log.Debug("repo.inventory.list.db_error", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, 0, m
	}

	log.Debug("repo.inventory.list.ok", zap.Duration("duration", time.Since(start)))
	return rows, total, nil
}

func (r *repository) Create(ctx context.Context, inventory *domain.Inventory) (*domain.Inventory, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	if err := r.db.WithContext(ctx).Create(inventory).Error; err != nil {
		m := apperror.MapDBError("repo.inventory.create", err)
		log.Debug("repo.inventory.create.db_fail", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, m
	}

	if err := r.cache.del(ctx, r.cache.keyID(inventory.ProductID)); err != nil {
		log.Warn("repo.inventory.create.cache.del_error", zap.Error(err))
	}

	log.Debug("repo.inventory.create.ok", zap.Uint("inventory_id", inventory.ID), zap.Duration("duration", time.Since(start)))
	return inventory, nil
}

// delta สามารถเป็นค่า + หรือ - ได้
func (r *repository) UpdateQuantity(ctx context.Context, pID uint, delta int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		log := ctxlog.From(ctx)
		start := time.Now()

		var inventory domain.Inventory

		// SELECT FOR UPDATE ล็อกแถวที่ถูกเลือกเพื่อป้องกันไม่ให้ข้อมูลถูกลบหร่ือแก้ไข
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id", pID).
			First(&inventory).Error; err != nil {

			m := apperror.MapDBError("repo.inventory.updateQuantity", err)
			log.Debug("repo.inventory.updatequantity.findinventory.db_error", zap.Error(err), zap.Duration("duration", time.Since(start)))
			return m
		}

		if err := tx.Model(&inventory).Where("product_id = ?", pID).
			Update("quantity", gorm.Expr("quantity + ?", delta)).Error; err != nil {

			m := apperror.MapDBError("repo.inventory.updatequantity", err)
			log.Debug("repo.inventory.updatequantity.db_error", zap.Error(err), zap.Duration("duration", time.Since(start)))
			return m
		}

		log.Debug("repo.inventory.updatequantity.ok", zap.Uint("product_id", pID), zap.Duration("duration", time.Since(start)))
		return nil
	})
}

func (r *repository) Delete(ctx context.Context, pID uint) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	if err := r.db.WithContext(ctx).Where("product_id = ?", pID).Delete(&domain.Inventory{}).Error; err != nil {
		m := apperror.MapDBError("repo.inventory.delete", err)
		log.Debug("repo.inventory.delete.db_error", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return m
	}

	// del cache
	if err := r.cache.del(ctx, r.cache.keyID(pID)); err != nil {
		log.Warn("repo.inventory.delete.cache_err", zap.Error(err))
	}

	log.Debug("repo.inventory.delete.ok", zap.Uint("product_id", pID), zap.Duration("duration", time.Since(start)))
	return nil
}
