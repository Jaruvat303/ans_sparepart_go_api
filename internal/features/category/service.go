package category

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/utils"
	"errors"

	"context"

	"go.uber.org/zap"
)

type Service interface {
	CreateCategory(ctx context.Context, req CategoryRequest) (*Item, error)
	GetCategoryByID(ctx context.Context, id uint) (*Item, error)
	GetCategoryByName(ctx context.Context, req CategoryRequest) (*Item, error)
	UpdateCategory(ctx context.Context, id uint, req CategoryRequest) (*Item, error)
	DeleteCategory(ctx context.Context, id uint) error
	List(ctx context.Context, q ListQuery) (*ListOutput, error)
}

type service struct {
	categoryRepo Repository
}

func NewService(categoryRepo Repository) Service {
	return &service{
		categoryRepo: categoryRepo,
	}
}

// CreateCategory creates a new product category
func (i *service) CreateCategory(ctx context.Context, req CategoryRequest) (*Item, error) {
	log := ctxlog.From(ctx)

	// Validate category name
	if req.Name == "" {
		return nil, apperror.ErrInvalidInput
	}

	// Check if category already exists (optional: check by name)
	cat, err := i.categoryRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}
	if cat != nil {
		return nil, apperror.ErrConflict
	}

	category := &domain.Category{
		Name: req.Name,
	}
	if err := i.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	// Log business event
	log.Info("category.created", zap.Uint("category_id", category.ID))

	return &Item{ID: category.ID, Name: category.Name}, nil
}

// GetCategoryByID retrieves a category by ID
func (i *service) GetCategoryByID(ctx context.Context, categoryID uint) (*Item, error) {

	category, err := i.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	return &Item{ID: category.ID, Name: category.Name}, nil
}

// GetCategoryByName retrieves a category by name
func (i *service) GetCategoryByName(ctx context.Context, req CategoryRequest) (*Item, error) {
	// Sanitize category name
	req.Name = utils.SanitizeString(req.Name)

	category, err := i.categoryRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &Item{ID: category.ID, Name: category.Name}, nil
}

// UpdateCategory updates a category's information
func (i *service) UpdateCategory(ctx context.Context, categoryID uint, req CategoryRequest) (*Item, error) {
	log := ctxlog.From(ctx)

	// validation and normalize categoryname
	req.Name = utils.SanitizeString(req.Name)
	if req.Name == "" {
		return nil, apperror.ErrInvalidInput
	}

	// Check if category already exists (optional: check by name)
	cat, err := i.categoryRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}
	if cat != nil {
		return nil, apperror.ErrConflict
	}

	// Retrieve existing category by ID
	category, err := i.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	// setting new data
	category.Name = req.Name
	// Save changes
	if err := i.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	log.Info("category.updated", zap.Uint("category_id", category.ID))
	return &Item{ID: category.ID, Name: category.Name}, nil
}

// DeleteCategory deletes a category if it has no associated products
func (i *service) DeleteCategory(ctx context.Context, categoryID uint) error {
	log := ctxlog.From(ctx)

	// Check if category exists
	category, err := i.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}

	// Delete category
	if err := i.categoryRepo.Delete(ctx, categoryID); err != nil {
		return err
	}

	log.Info("category_deleted", zap.Uint("category_id", category.ID))

	return nil
}

// List
func (i *service) List(ctx context.Context, q ListQuery) (*ListOutput, error) {
	limit, offset := utils.NormalizePagination(q.Limit, q.Offset)

	rows, total, err := i.categoryRepo.List(ctx, ListQuery{
		Search: q.Search,
		Limit:  limit,
		Offset: offset,
		Sort:   q.Sort,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*Item, 0, len(rows))
	for _, category := range rows {
		items = append(items, &Item{
			ID:   category.ID,
			Name: category.Name,
		})
	}

	return &ListOutput{Items: items, Total: total}, nil
}
