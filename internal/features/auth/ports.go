package auth

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/infra/jwtx"
	"context"
	"time"
)

type AuthRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

// TokenIssuer interface implement by auth/jwt
type TokenIssuer interface {
	GenerateToken(userID uint, username, role string) (token string, jwtID string, err error)
	ValidateToken(token string) (*jwtx.Claims, error)
	GetExpiry(token string) (time.Duration, error)
	BlacklistToken(ctx context.Context, jwtID string, exp time.Duration) error
}


