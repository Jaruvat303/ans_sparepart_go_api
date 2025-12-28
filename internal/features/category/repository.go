package category

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

// Categort Repository interface
type Repository interface {
	Create(ctx context.Context, category *domain.Category) error
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id uint) error
	
	List(ctx context.Context, q ListQuery) ([]*domain.Category, int64, error)
	GetByID(ctx context.Context, id uint) (*domain.Category, error)
	GetByName(ctx context.Context, name string) (*domain.Category, error)
}

type repository struct {
	db    *gorm.DB
	cache *cacheLayer
}

func NewRepository(db *gorm.DB, rdb *redis.Client, exp time.Duration) Repository {
	return &repository{
		db:    db,
		cache: newCache(rdb, exp),
	}
}

func (r *repository) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// check cache
	if cache, ok, err := r.cache.getByKey(ctx, r.cache.keyID(id)); err == nil && ok {
		log.Debug("repo.category.getByID.cache_hit", zap.Uint("id", id), zap.Duration("duration", time.Since(start)))
		return cache, nil
	} else if err != nil {
		log.Warn("repo.category.getByID.cache_error", zap.Uint("category_id", id), zap.Error(err))
	}

	// DB
	var c domain.Category
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		m := apperror.MapDBError("repo.category.getByID", err)
		log.Debug("repo.category.getByID.db_fail", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyID(c.ID), &c); err != nil {
		log.Debug("repo.category.getByID.cache_set_fail", zap.Error(err))
	}

	log.Debug("repo.category.getByID.ok", zap.Uint("id", id), zap.Duration("duration", time.Since(start)))
	return &c, nil
}

func (r *repository) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// Cache chack
	if cache, ok, err := r.cache.getByKey(ctx, r.cache.keyName(name)); err == nil && ok {
		log.Debug("repo.category.getbyname.cache_hit", zap.Uint("category_id", cache.ID), zap.Duration("duration", time.Since(start)))
		return cache, nil
	} else if err != nil {
		log.Warn("repo.category.getbyname.cache_error", zap.String("category_name", name), zap.Error(err))
	}

	// DB
	var c domain.Category
	if err := r.db.WithContext(ctx).First(&c, "name = ?", name).Error; err != nil {
		m := apperror.MapDBError("repo.category.getbyname", err)
		log.Debug("repo.category.getbyname.db_error", zap.String("category_name", name), zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyName(name), &c); err != nil {
		log.Warn("repo.category.getbyname.set_cache.error", zap.String("category_name", name), zap.Error(err))
	}

	log.Debug("repo.category.getbyname.ok", zap.String("category_name", name), zap.Duration("duration", time.Since(start)))
	return &c, nil
}

func (r *repository) List(ctx context.Context, q ListQuery) ([]*domain.Category, int64, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// create Query Builder
	tx := r.db.WithContext(ctx).Model(&domain.Category{})

	if q.Search == "" {
		tx = tx.Where("name =  ILIKE  ", "%"+q.Search+"%")
	}

	// cout รวม
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		m := apperror.MapDBError("repo.category.list.count", err)
		log.Debug("repo.category.list.count.fail", zap.Error(err))
		return nil, 0, m
	}

	if q.Sort == "" {
		tx = tx.Order(q.Sort)
	} else {
		tx = tx.Order("created_at DESC")
	}
	if q.Offset != 0 {
		tx = tx.Offset(q.Offset)
	}
	if q.Limit != 0 {
		tx = tx.Limit(q.Limit)
	}

	var rows []*domain.Category
	if err := tx.Find(&rows).Error; err != nil {
		m := apperror.MapDBError("repo.category.list", err)
		log.Debug("repo.category.list.db_error", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, 0, m
	}

	log.Debug("repo.category.list.ok", zap.Duration("duration", time.Since(start)))
	return rows, total, nil

}

func (r *repository) Create(ctx context.Context, category *domain.Category) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	// DB
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		m := apperror.MapDBError("repo.category.create", err)
		log.Debug("repo.category.create.db_fail", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return m
	}

	// cache
	if err := r.cache.del(ctx, r.cache.keyID(category.ID)); err != nil {
		log.Warn("repo.category.create.cache.del_fail", zap.Error(err))
	}

	log.Debug("repo.category.create.ok", zap.Uint("id", category.ID), zap.Duration("duration", time.Since(start)))
	return nil

}

func (r *repository) Update(ctx context.Context, category *domain.Category) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	// DB
	if err := r.db.WithContext(ctx).Save(category).Error; err != nil {
		m := apperror.MapDBError("repo.category.update", err)
		log.Debug("repo.category.update.fail", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return m
	}

	// del cache
	if err := r.cache.del(ctx, r.cache.keyID(category.ID)); err != nil {
		log.Warn("repo.category.update.cache.del_fail", zap.Error(err))
	}

	log.Debug("repo.category.update.ok", zap.Uint("id", category.ID), zap.Duration("duration", time.Since(start)))
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	// DB
	if err := r.db.WithContext(ctx).Delete(domain.Category{}, id).Error; err != nil {
		m := apperror.MapDBError("repo.category.delete", err)
		log.Debug("repo.category.delete.fail", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return m
	}

	// del cache
	if err := r.cache.del(ctx, r.cache.keyID(id)); err != nil {
		log.Warn("repo.category.delete.cache.fail", zap.Error(err))
	}

	return nil
}
