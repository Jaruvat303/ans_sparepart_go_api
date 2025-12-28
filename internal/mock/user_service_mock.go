package mocks

import (
	"ans-spareparts-api/internal/features/user"
	"context"

	"github.com/stretchr/testify/mock"
)

type UserService struct {
	mock.Mock
}

func NewUserService() *UserService {
	return &UserService{}
}

func (m *UserService) GetUserProfile(ctx context.Context, userID uint) (*user.Item, error) {
	args := m.Called(ctx, userID)
	if value, ok := args.Get(0).(*user.Item); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserService) DeleteUser(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
