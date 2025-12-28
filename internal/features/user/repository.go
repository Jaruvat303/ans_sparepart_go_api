package user

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

// UserRepository repesents the user repository contract
type Repository interface {
	// -- UserProfile Method --
	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	// Update updates an existsing user's data
	Update(ctx context.Context, user *domain.User) error
	// Delete set field deleteAt a user from the database by ID.
	Delete(ctx context.Context, id uint) error

	// -- Auth Method --
	// GetByUsername retriecves a user by their username
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	// GetByEmail retriecves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	// Create insert a new user into the database.
	Create(ctx context.Context, user *domain.User) error
}

// repository struct
type repository struct {
	db    *gorm.DB
	cache *cacheLayer
}

// NewRepository Constructor
func NewRepository(db *gorm.DB, rdb *redis.Client, ttl time.Duration) Repository {
	return &repository{
		db:    db,
		cache: newCache(rdb, ttl),
	}
}

func (r *repository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// check cache
	if cache, ok, err := r.cache.getByKey(ctx, r.cache.keyByID(id)); err == nil && ok {
		log.Debug("repo.user.getbyid.cache_hit", zap.Uint("user_id", id), zap.Duration("duration", time.Since(start)))
		return cache, nil
	} else if err != nil {
		log.Warn("repo.user.getbyid.cache_err", zap.Error(err))
	}

	// DB
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		m := apperror.MapDBError("repo.user.getbyid", err)
		log.Debug("repo.user.getbyid.db_error", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyByID(id), &user); err != nil {
		log.Warn("repo.user.getbyid.cache_set_err", zap.Error(err))
	}

	log.Debug("repo.user.getbyid.ok", zap.Uint("user_id", id), zap.Duration("duration", time.Since(start)))
	return &user, nil
}

func (r *repository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// cache check
	if cache, ok, err := r.cache.getByKey(ctx, r.cache.keyByUsername(username)); err == nil && ok {
		log.Debug("repo.user.getbyusername.cache_hit", zap.Duration("duration", time.Since(start)))
		return cache, nil
	} else if err != nil {
		log.Warn("repo.user.getbyusername.cache_hit", zap.Error(err))
	}

	// DB
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "username = ? ", username).Error; err != nil {
		m := apperror.MapDBError("repo.user.getbyusername", err)
		log.Debug("repo.user.getbyusername.db_error", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyByUsername(username), &user); err != nil {
		log.Warn("repo.user.getbyusername.set_cache_error", zap.Error(err))
	}

	log.Debug("repo.user.getbyusername.ok", zap.Duration("duration", time.Since(start)))
	return &user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	log := ctxlog.From(ctx)
	start := time.Now()

	// cache check
	if cache, ok, err := r.cache.getByKey(ctx, r.cache.keyByEmail(email)); err == nil && ok {
		log.Debug("repo.user.getbyemail.cache_hit", zap.Duration("duration", time.Since(start)))
		return cache, nil
	} else if err != nil {
		log.Warn("repo.user.getbyemail.cache_err", zap.Error(err))
	}

	// DB
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		m := apperror.MapDBError("repo.user.getbyemail", err)
		log.Debug("repo.user.getbyemail.db_err", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return nil, m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyByEmail(email), &user); err != nil {
		log.Warn("repo.user.getbyemail.set_cache_err", zap.Error(err))
	}

	log.Debug("repo.user.getbyemail.ok", zap.Duration("duration", time.Since(start)))
	return &user, nil
}

func (r *repository) Create(ctx context.Context, user *domain.User) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	// DB
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		m := apperror.MapDBError("repo.user.create", err)
		log.Debug("repo.user.create.db_error", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return m
	}

	// set cache
	if err := r.cache.set(ctx, r.cache.keyByID(user.ID), user); err != nil {
		log.Warn("repo.user.create.set_cache.fail", zap.Error(err))
	}

	log.Debug("repo.user.create.ok", zap.Uint("user_id", user.ID), zap.Duration("duration", time.Since(start)))
	return nil
}

func (r *repository) Update(ctx context.Context, user *domain.User) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		m := apperror.MapDBError("repo.user.update", err)
		log.Debug("repo.user.update.db_err", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return m
	}

	// del cache
	if err := r.cache.del(ctx, r.cache.keyByID(user.ID)); err != nil {
		log.Warn("repo.user.update.del_cache.fail", zap.Error(err))
	}

	log.Debug("repo.user.update.ok", zap.Uint("user_id", user.ID), zap.Duration("duration", time.Since(start)))
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	log := ctxlog.From(ctx)
	start := time.Now()

	// DB
	if err := r.db.WithContext(ctx).Delete(&domain.User{}, id).Error; err != nil {
		m := apperror.MapDBError("repo.user.delte", err)
		log.Debug("repo.user.delete.db_err", zap.Error(err), zap.Duration("duration", time.Since(start)))
		return m
	}

	if err := r.cache.del(ctx, r.cache.keyByID(id)); err != nil {
		log.Warn("repo.user.delete.del_cache.fail", zap.Error(err))
	}

	log.Debug("repo.user.delete.ok", zap.Uint("user_id", id), zap.Duration("duraion", time.Since(start)))
	return nil
}
