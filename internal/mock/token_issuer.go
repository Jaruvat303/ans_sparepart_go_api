package mocks

import (
	"ans-spareparts-api/internal/infra/jwtx"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type TokenIssuer struct {
	mock.Mock
}

func NewTokenIssuer() *TokenIssuer {
	return &TokenIssuer{}
}

func (m *TokenIssuer) GenerateToken(uID uint, username, role string) (string, string, error) {
	args := m.Called(uID, username, role)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *TokenIssuer) ValidateToken(token string) (*jwtx.Claims, error) {
	args := m.Called(token)
	if c, ok := args.Get(0).(*jwtx.Claims); ok {
		return c, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *TokenIssuer) GetExpiry(token string) (time.Duration, error) {
	args := m.Called(token)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *TokenIssuer) BlacklistToken(ctx context.Context, jwtID string, ttl time.Duration) error {
	return m.Called(ctx, jwtID, ttl).Error(0)
}
