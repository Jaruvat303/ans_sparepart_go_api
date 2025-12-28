package auth

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/infra/hash"
	"ans-spareparts-api/internal/infra/httpx/ctxlog"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/utils"
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type Service interface {
	Register(ctx context.Context, in RegisterInput) (*domain.User, error)
	Login(ctx context.Context, in LoginInput) (*domain.User, string, error)
	Logout(ctx context.Context, token string) error
}

type service struct {
	authRepo AuthRepository
	hash     hash.Hasher
	tokens   TokenIssuer

	DefaultRole string
}

func NewService(userRepo AuthRepository, tokens TokenIssuer, hash hash.Hasher, defaultRole string) Service {
	if defaultRole == "" {
		defaultRole = "cashier"
	}

	return &service{
		authRepo:    userRepo,
		tokens:      tokens,
		hash:        hash,
		DefaultRole: defaultRole,
	}
}

// --- helper ---
func sanitizeRegister(in RegisterInput) error {
	if strings.TrimSpace(in.Email) == "" ||
		strings.TrimSpace(in.Username) == "" ||
		strings.TrimSpace(in.Password) == "" {
		return apperror.ErrInvalidInput
	}
	return nil
}

func sanitizeLogin(in LoginInput) error {
	if strings.TrimSpace(in.Username) == "" ||
		strings.TrimSpace(in.Password) == "" {
		return apperror.ErrInvalidInput
	}
	return nil
}

// --- Method ---

// Register: normalized username/email -> hash -> created -> clear password -> return
func (i *service) Register(ctx context.Context, in RegisterInput) (*domain.User, error) {
	log := ctxlog.From(ctx)

	// null check
	if err := sanitizeRegister(in); err != nil {
		log.Warn("interactor.auth.register.validate_input", zap.Error(err))
		return nil, err
	}

	username := strings.TrimSpace(in.Username)
	email := strings.ToLower(strings.TrimSpace(in.Email))
	role := strings.TrimSpace(in.Role)
	if role == "" {
		role = i.DefaultRole
	}

	// email format check
	if ok := utils.IsValidEmail(email); !ok {
		log.Warn("interactor.auth.register.invalid_email_format")
		return nil, apperror.ErrInvalidInput
	}

	// password format check
	if err := utils.VerifyPasswordStrength(in.Password); err != nil {
		log.Warn("interactor.auth.register.invalid_password_format")
		return nil, apperror.ErrInvalidInput
	}

	// Check if user already exists
	existingUser, err := i.authRepo.GetByUsername(ctx, username)
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}
	if existingUser != nil {
		log.Warn("interactor.auth.register.username_already_exist")
		return nil, apperror.ErrConflict
	}

	existingUser, err = i.authRepo.GetByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}
	if existingUser != nil {
		log.Warn("interactor.auth.register.email.already_exist")
		return nil, apperror.ErrConflict
	}

	// Hash password
	hashedPassword, err := i.hash.HashPassword(in.Password)
	if err != nil {
		log.Warn("interactor.auth.register.hash_password.fail", zap.Error(err))
		return nil, apperror.ErrInternalServer
	}

	// Create user
	user := &domain.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     role,
		IsActive: true,
	}

	if err := i.authRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Return user without password
	user.Password = ""
	log.Info("user.registered", zap.Uint("user_id", user.ID), zap.String("username", username), zap.String("role", user.Role))
	return user, nil
}

func (i *service) Login(ctx context.Context, in LoginInput) (*domain.User, string, error) {
	log := ctxlog.From(ctx)

	if err := sanitizeLogin(in); err != nil {
		log.Warn("interactor.auth.login.validate_input", zap.Error(err))
		return nil, "", err
	}
	username := strings.TrimSpace(in.Username)

	// Find user by username
	user, err := i.authRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", err
	}

	// Check if user is active
	if !user.IsActive {
		log.Warn("interactor.auth.login.GetByUsername.user_not_active", zap.Uint("user_id", user.ID))
		return nil, "", apperror.ErrUserForbidden
	}

	// Verify password
	if err := i.hash.CompareHashAndPassword(in.Password, user.Password); err != nil {
		log.Warn("interactor.auth.login.CompareHashAndPassword.fail", zap.Error(err))
		return nil, "", apperror.ErrUnauthorized
	}

	// Generate token
	token, _, err := i.tokens.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		log.Warn("interactor.auth.login.generate_token.fail", zap.Error(err))
		return nil, "", apperror.ErrInternalServer
	}

	// reset password
	user.Password = ""
	log.Info("user.login", zap.Uint("user_id", user.ID), zap.String("username", username))

	return user, token, nil
}

func (i *service) Logout(ctx context.Context, token string) error {
	log := ctxlog.From(ctx)

	if strings.TrimSpace(token) == "" {
		log.Warn("interactor.auth.logout.missing_token")
		return nil
	}

	ttl, err := i.tokens.GetExpiry(token)
	if err != nil || ttl <= 0 {
		// ถ้า Token ผิด format หรือ parse ไม่ได้ ก็ถือว่า logout สำเร็จไปเลย (ไม่ต้องทำไรต่อ)
		return nil
	}

	// Validate token to get expiry time
	claims, err := i.tokens.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", apperror.ErrInvalidToken)
	}

	// Blacklist the token
	if err := i.tokens.BlacklistToken(ctx, claims.ID, ttl); err != nil {
		// ไม่ fail ทั้ง request — แค่ log เว้นแต่ธุรกิจบังคับ
		log.Warn("logout.blacklist_failed", zap.Error(err))
	}

	log.Info("user.logout", zap.Uint("user_id", claims.UserID), zap.String("JWTID", claims.ID))
	return nil
}
