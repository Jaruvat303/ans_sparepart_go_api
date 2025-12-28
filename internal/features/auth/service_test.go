package auth_test

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/auth"
	"ans-spareparts-api/internal/infra/jwtx"
	mocks "ans-spareparts-api/internal/mock"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/testutil/fixtures"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestSuite struct {
	MockUserRepo *mocks.UserRepository
	MockHash     *mocks.Hasher
	MockToken    *mocks.TokenIssuer
	Service      auth.Service
	Ctx          context.Context
}

func NewTestSuite() *TestSuite {
	return &TestSuite{}
}

func (ts *TestSuite) SetupTestSuite(t *testing.T) {
	ts.MockHash = mocks.NewHasher()
	ts.MockToken = mocks.NewTokenIssuer()
	ts.MockUserRepo = mocks.NewMockUserRepository()
	ts.Service = auth.NewService(ts.MockUserRepo, ts.MockToken, ts.MockHash, "cashier")
	ts.Ctx = context.Background()

	t.Cleanup(func() {
		ts.MockHash.AssertExpectations(t)
		ts.MockToken.AssertExpectations(t)
		ts.MockUserRepo.AssertExpectations(t)
	})
}

// --- Register ---
func TestAuthService_Register(t *testing.T) {
	mockUser := fixtures.ValidUser()
	input := &auth.RegisterInput{
		Username: mockUser.Username,
		Email:    mockUser.Email,
		Password: "P@ssword1234",
	}

	tests := []struct {
		name      string
		in        auth.RegisterInput
		setup     func(ts *TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *domain.User)
	}{
		{
			name: "success with default role",
			in:   *input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, input.Username).Return(nil, apperror.ErrNotFound).Once()
				ts.MockUserRepo.On("GetByEmail", ts.Ctx, input.Email).Return(nil, apperror.ErrNotFound).Once()
				ts.MockHash.On("HashPassword", input.Password).Return("hashed", nil).Once()
				ts.MockUserRepo.On("Create", ts.Ctx, mock.MatchedBy(func(u *domain.User) bool {
					assert.Equal(t, mockUser.Username, u.Username)
					assert.Equal(t, mockUser.Email, u.Email)
					assert.Equal(t, mockUser.Role, u.Role)
					u.ID = 1 // จำลอง ID หลัง Created
					return true
				})).Return(nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, u *domain.User) {
				// assert.Equal(t, uint(1), u.ID)
				assert.Empty(t, u.Password)
				assert.Equal(t, "cashier", u.Role)
			},
		},
		{
			name: "error: username exist",
			in:   *input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, input.Username).Return(mockUser, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrConflict)
			},
			validate: func(t *testing.T, u *domain.User) {
				assert.Nil(t, u)
			},
		},
		{
			name: "error: email exists",
			in:   *input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, input.Username).Return(nil, apperror.ErrNotFound).Once()
				ts.MockUserRepo.On("GetByEmail", ts.Ctx, input.Email).Return(mockUser, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrConflict)
			},
			validate: func(t *testing.T, u *domain.User) {
				assert.Nil(t, u)
			},
		},
		{
			name: "error: hash fails",
			in:   *input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, input.Username).Return(nil, apperror.ErrNotFound)
				ts.MockUserRepo.On("GetByEmail", ts.Ctx, input.Email).Return(nil, apperror.ErrNotFound)
				ts.MockHash.On("HashPassword", "P@ssword1234").Return("", errors.New("oops"))
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrInternalServer)
			},
			validate: func(t *testing.T, u *domain.User) {
				assert.Nil(t, u)
			},
		},
		{
			name: "error: create fails",
			in:   *input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, input.Username).Return(nil, apperror.ErrNotFound).Once()
				ts.MockUserRepo.On("GetByEmail", ts.Ctx, input.Email).Return(nil, apperror.ErrNotFound).Once()
				ts.MockHash.On("HashPassword", "P@ssword1234").Return("hashed", nil).Once()
				ts.MockUserRepo.On("Create", ts.Ctx, mock.MatchedBy(func(u *domain.User) bool {
					assert.Equal(t, mockUser.Username, u.Username)
					assert.Equal(t, mockUser.Email, u.Email)
					assert.Equal(t, mockUser.Role, u.Role)
					u.ID = 1 // จำลอง ID หลัง Created
					return true
				})).Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, u *domain.User) {
				assert.Nil(t, u)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTestSuite(t)

			test.setup(ts)
			got, err := ts.Service.Register(ts.Ctx, test.in)

			test.assertErr(t, err)
			test.validate(t, got)

		})
	}
}

// --- Login --
func TestAuthService_Login(t *testing.T) {
	mockUser := fixtures.ValidUser()
	input := auth.LoginInput{
		Username: mockUser.Username,
		Password: mockUser.Password,
	}
	tests := []struct {
		name      string
		in        auth.LoginInput
		setup     func(ts *TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *domain.User, string)
	}{
		{
			name: "success",
			in:   input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, input.Username).Return(mockUser, nil)
				ts.MockHash.On("CompareHashAndPassword", mockUser.Password, "P@ssword1234").Return(nil)
				ts.MockToken.On("GenerateToken", mockUser.ID, mockUser.Username, mockUser.Role).Return("token1234", "jwtID", nil)
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, u *domain.User, token string) {
				assert.NotNil(t, u)
				assert.Equal(t, input.Username, u.Username)
				assert.Equal(t, "token1234", token)
				assert.Empty(t, u.Password)
			},
		},
		{
			name: "user not found",
			in:   input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, input.Username).Return(nil, apperror.ErrNotFound)
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrNotFound)
			},
			validate: func(t *testing.T, u *domain.User, s string) {
				assert.Nil(t, u)
			},
		},
		{
			name: "user inactive",
			in:   input,
			setup: func(ts *TestSuite) {
				newUser := fixtures.ValidUser()
				newUser.IsActive = false
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, mockUser.Username).Return(newUser, nil)
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrUserForbidden)
			},
			validate: func(t *testing.T, u *domain.User, s string) {
				assert.Nil(t, u)
			},
		},
		{
			name: "wrong password",
			in:   input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, mockUser.Username).Return(mockUser, nil)
				ts.MockHash.On("CompareHashAndPassword", input.Password, mockUser.Password).Return(apperror.ErrInvalidInput)

			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrUnauthorized)
			},
			validate: func(t *testing.T, u *domain.User, s string) {
				assert.Nil(t, u)
			},
		},
		{
			name: "genterate token fails",
			in:   input,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByUsername", ts.Ctx, input.Username).Return(mockUser, nil)
				ts.MockHash.On("CompareHashAndPassword", input.Password, mockUser.Password).Return(nil)
				ts.MockToken.On("GenerateToken", mockUser.ID, mockUser.Username, mockUser.Role).Return("", "", apperror.ErrInvalidInput)

			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrInternalServer)
			},
			validate: func(t *testing.T, u *domain.User, s string) {
				assert.Nil(t, u)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTestSuite(t)

			test.setup(ts)
			user, token, err := ts.Service.Login(ts.Ctx, test.in)

			test.assertErr(t, err)
			test.validate(t, user, token)
		})
	}
}

// // --- Logout ---
func TestAuthService_Logout(t *testing.T) {
	validToken := "valid.jwt.token"
	mockClaims := &jwtx.Claims{
		UserID:   1,
		Username: "Test",
		Role:     "cachier",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	validTTL := time.Hour

	tests := []struct {
		name      string
		token     string
		setup     func(*TestSuite)
		assertErr func(*testing.T, error)
	}{
		{
			name:  "Success_Token_Valid_Blacklisted",
			token: validToken,
			setup: func(ts *TestSuite) {
				ts.MockToken.On("GetExpiry", validToken).Return(validTTL, nil).Once()
				ts.MockToken.On("ValidateToken", validToken).Return(mockClaims, nil).Once()
				ts.MockToken.On("BlacklistToken", ts.Ctx, mockClaims.ID, validTTL).Return(nil).Once()

			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "Success_Blacklust_Failed_ButNoError",
			token: validToken,
			setup: func(ts *TestSuite) {
				ts.MockToken.On("GetExpiry", validToken).Return(validTTL, nil).Once()
				ts.MockToken.On("ValidateToken", validToken).Return(mockClaims, nil).Once()
				ts.MockToken.On("BlacklistToken", ts.Ctx, mockClaims.ID, validTTL).
					Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "Success_EmptyOrWhitespace_Token",
			token: " ",
			setup: func(ts *TestSuite) {

			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "Success_Token_Expired_Or_InvalidTTL",
			token: validToken,
			setup: func(ts *TestSuite) {
				// GetExpiry Return TTL <= 0
				ts.MockToken.On("GetExpiry",validToken).Return(time.Duration(0),nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t,err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTestSuite(t)

			test.setup(ts)
			err := ts.Service.Logout(ts.Ctx, test.token)

			test.assertErr(t, err)
		})
	}

}
