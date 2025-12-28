package mocks

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/auth"
	"context"

	"github.com/stretchr/testify/mock"
)

type AuthService struct {
	mock.Mock
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (m *AuthService) Register(ctx context.Context, input auth.RegisterInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	if value, ok := args.Get(0).(*domain.User); ok {
		return value, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *AuthService) Login(ctx context.Context, input auth.LoginInput) (*domain.User, string, error) {
	args := m.Called(ctx, input)
	if value, ok := args.Get(0).(*domain.User); ok {
		return value, args.String(1), args.Error(2)
	}
	return nil, args.String(1), args.Error(2)
}

func (m *AuthService) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}
