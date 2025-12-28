package product

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/pkg/apperror"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository interface {
	GetByID(ctx context.Context, id uint) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	List(ctx context.Context, q ListQuery) ([]*domain.Product, int64, error)

	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id uint) error
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

func (r *repository) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// cashe
	if cache, ok, err := r.cache.getByKey(ctx, r.cache.keyByID(id)); err == nil && ok {
		log.Debug("repo.product.GetByID.cache_hit", zap.Uint("id", id), zap.Duration("duration", time.Since(start)))
		return cache, nil
	} else if err != nil {
		log.Warn("repo.product.getByID.cashe_err", zap.Uint("product_id", id), zap.Error(err))
	}

	// DB
	var p domain.Product
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		m := apperror.MapDBError("repo.product.getByID", err)
		log.Debug("repo.product.getByID.db_fail", zap.Error(err), zap.Duration("duraion", time.Since(start)))
		return nil, m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyByID(id), &p); err != nil {
		log.Warn("repo.product.getByID.cache_set_err", zap.Error(err))
	}

	log.Debug("repo.product.getByID.ok", zap.Duration("duration", time.Since(start)))
	return &p, nil
}


func (r *repository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	if p, ok, err := r.cache.getByKey(ctx, r.cache.keyBySKU(sku)); err == nil && ok {
		log.Debug("repo.product.getBySKU.cache_hit", zap.String("sku", sku), zap.Duration("duration", time.Since(start)))
		return p, nil
	} else if err != nil {
		log.Warn("repo.product.getBySKU.cache_err", zap.Error(err))
	}

	var p domain.Product
	if err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&p).Error; err != nil {
		m := apperror.MapDBError("repo.product.getBySKU", err)
		log.Debug("repo.product.getBySKU.db_fail", zap.Error(err))
		return nil, m
	}

	// เขียน cache แบบสองคีย์ เพื่อรองรับ lookup ได่ทั้ง id และ sku
	if err := r.cache.set(ctx, r.cache.keyByID(p.ID), &p); err != nil {
		log.Warn("repo.product.getBySKU.cache_set_err", zap.Error(err), zap.Duration("duration", time.Since(start)))
	}
	if err := r.cache.set(ctx, r.cache.keyBySKU(sku), &p); err != nil {
		log.Warn("repo.product.getBySKU.cache_set_err", zap.Error(err), zap.Duration("duration", time.Since(start)))
	}

	log.Debug("repo.product.getBySKU.ok", zap.Duration("duration", time.Since(start)))
	return &p, nil
}

func (r *repository) List(ctx context.Context, q ListQuery) ([]*domain.Product, int64, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// Create Query Builder Session
	tx := r.db.WithContext(ctx).Model(&domain.Product{})

	if q.Search != "" {
		tx = tx.Where("name ILIKE ? OR sku ILIKE ?", "%"+q.Search+"%", "%"+q.Search+"%")
	}

	// count รวม
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		m := apperror.MapDBError("repo.product.list.count", err)
		log.Debug("repo.product.list.count_fail", zap.Error(err))
		return nil, 0, m
	}

	// sort + page
	if q.Sort != "" {
		tx = tx.Order(q.Sort)
	} else {
		tx = tx.Order("created_at DESC")
	}
	if q.Limit > 0 {
		tx = tx.Limit(q.Limit)
	}
	if q.Offset > 0 {
		tx = tx.Offset(q.Offset)
	}

	// Find with Query Builder Session
	var rows []*domain.Product
	if err := tx.Find(&rows).Error; err != nil {
		m := apperror.MapDBError("repo.product.list.find", err)
		log.Debug("repo.product.list.find_fail", zap.Error(err))
		return nil, 0, m
	}

	log.Debug("repo.product.list.ok", zap.Int("n", len(rows)), zap.Int64("total", total), zap.Duration("duration", time.Since(start)))
	return rows, total, nil
}

func (r *repository) Create(ctx context.Context, p *domain.Product) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		m := apperror.MapDBError("repo.product.create", err)
		log.Debug("repo.product.create.fail", zap.Error(err))
		return m
	}

	// invalidate cache for keys that might affect reads
	if err := r.cache.del(ctx, r.cache.keyByID(p.ID), r.cache.keyBySKU(p.SKU)); err != nil {
		log.Warn("repo.product.create.cache.del.error", zap.Error(err))
	}

	log.Info("repo.product.create.ok", zap.Uint("id", p.ID), zap.Duration("duration", time.Since(start)))
	return nil
}

func (r *repository) Update(ctx context.Context, p *domain.Product) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	if err := r.db.WithContext(ctx).Save(p).Error; err != nil {
		m := apperror.MapDBError("repo.product.update", err)
		log.Debug("repo.product.update.fail", zap.Error(err))
		return m
	}

	if err := r.cache.del(ctx, r.cache.keyByID(p.ID), r.cache.keyBySKU(p.SKU)); err != nil {
		log.Warn("repo.product.update.cache_del_fail", zap.Error(err))
	}

	log.Info("repo.product.update.ok", zap.Uint("id", p.ID), zap.Duration("duration", time.Since(start)))
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	var p domain.Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		m := apperror.MapDBError("repo.product.delete.getByID", err)
		log.Debug("repo.product.delte.getByID.fail", zap.Error(err))
		return m
	}

	if err := r.db.WithContext(ctx).Delete(&domain.Product{}, id).Error; err != nil {
		m := apperror.MapDBError("repo.product.delete", err)
		log.Debug("repo.product.delete.fail", zap.Error(err))
		return m
	}

	if err := r.cache.del(ctx, r.cache.keyByID(id), r.cache.keyBySKU(p.SKU)); err != nil {
		log.Warn("repo.product.delete.fail", zap.Error(err))
	}

	log.Debug("repo.product.delete.ok", zap.Uint("id", id), zap.Duration("duration", time.Since(start)))
	return nil
}
