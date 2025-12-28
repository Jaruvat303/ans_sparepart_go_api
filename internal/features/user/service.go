package user

import (
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/pkg/apperror"
	"context"

	"go.uber.org/zap"
)

// Service Interface
type Service interface {
	GetUserProfile(ctx context.Context, userID uint) (*Item, error)
	DeleteUser(ctx context.Context, userID uint) error
}

type service struct {
	userRepo Repository
}

func NewService(userRepo Repository) Service {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) GetUserProfile(ctx context.Context, userID uint) (*Item, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	item := &Item{
		ID:       userID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}

	return item, nil
}

func (s *service) DeleteUser(ctx context.Context, userID uint) error {
	log := ctxlog.From(ctx)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.ErrNotFound
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	log.Info("user profile deleted", zap.Uint("user_id", userID))
	return nil
}
