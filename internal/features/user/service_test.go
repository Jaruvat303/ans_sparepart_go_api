package user_test

import (
	"ans-spareparts-api/internal/features/user"
	mocks "ans-spareparts-api/internal/mock"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/testutil/fixtures"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestSuite struct {
	Service      user.Service
	MockUserRepo *mocks.UserRepository
	Ctx          context.Context
}

func newTestSuite() *TestSuite {
	return &TestSuite{}
}

func (ts *TestSuite) SetupTestSuite(t *testing.T) {
	ts.MockUserRepo = mocks.NewMockUserRepository()
	ts.Service = user.NewService(ts.MockUserRepo)
	ts.Ctx = context.Background()

	t.Cleanup(func() {
		ts.MockUserRepo.AssertExpectations(t)
	})
}

// GetUserProfile
func TestUserService_GetUserProfile(t *testing.T) {
	mockUser := fixtures.ValidUser()

	tests := []struct {
		name      string
		userID    uint
		setup     func(ts *TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *user.Item)
	}{
		{
			name:   "success_found_user",
			userID: 1,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByID", ts.Ctx, uint(1)).Return(mockUser, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *user.Item) {
				assert.NotNil(t, i)
				assert.Equal(t, mockUser.ID, i.ID)
				assert.Equal(t, mockUser.Username, i.Username)
			},
		},
		{
			name:   "error_user_not_found",
			userID: 99,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByID", ts.Ctx, uint(99)).Return(nil, apperror.ErrNotFound)
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrNotFound)
			},
			validate: func(t *testing.T, i *user.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name:   "error_internal_database_fail",
			userID: 50,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByID", ts.Ctx, uint(50)).Return(nil, apperror.ErrInternalServer)
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, apperror.ErrInternalServer)
				assert.NotErrorIs(t, err, apperror.ErrNotFound)
			},
			validate: func(t *testing.T, i *user.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := newTestSuite()
			ts.SetupTestSuite(t)

			test.setup(ts)
			item, err := ts.Service.GetUserProfile(ts.Ctx, test.userID)

			test.assertErr(t, err)
			test.validate(t, item)
		})
	}
}

// Delete Test
func TestUserService_Delete(t *testing.T) {
	mockuser := fixtures.ValidUser()
	tests := []struct {
		name      string
		userID    uint
		setup     func(ts *TestSuite)
		assertErr func(t *testing.T, err error)
	}{
		{
			name:   "success",
			userID: 1,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByID", ts.Ctx, uint(1)).Return(mockuser, nil).Once()
				ts.MockUserRepo.On("Delete", ts.Ctx, uint(1)).Return(nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:   "error_delete_db_fails",
			userID: 2,
			setup: func(ts *TestSuite) {
				ts.MockUserRepo.On("GetByID", ts.Ctx, uint(2)).Return(mockuser, nil).Once()
				ts.MockUserRepo.On("Delete", ts.Ctx, uint(2)).Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, apperror.ErrInternalServer)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := newTestSuite()
			ts.SetupTestSuite(t)

			test.setup(ts)
			err := ts.Service.DeleteUser(context.Background(), test.userID)

			test.assertErr(t, err)
		})
	}
}

// Test Update
// func TestUserService_Update(t *testing.T) {
// 	mockRepo := mocks.NewMockUserRepository()
// 	userID := uint(1)
// 	newSV := user.NewService(mockRepo)

// 	tests := []struct {
// 		name      string
// 		input     user.UpdateInput
// 		setup     func(ts *TestSuite)
// 		assertErr func(*testing.T, error)
// 		validate  func(*testing.T, *user.Item)
// 	}{
// 		{
// 			name: "success_update_alllFields",
// 			input: user.UpdateInput{
// 				Email:    testutil.PTRHelper("test123@mail.com"),
// 				Password: testutil.PTRHelper("J@rovat1234"),
// 			},
// 			setup: func(ts *TestSuite) {
// 				u := fixtures.ValidUser()
// 				m.On("GetByID", ts.Ctx, userID).Return(u, nil).Once()
// 				m.On("GetByEmail", ts.Ctx, "test123@mail.com").Return(nil, apperror.ErrNotFound).Once()
// 				m.On("Update", ts.Ctx, u).Return(nil).Once()
// 			},
// 			assertErr: func(t *testing.T, err error) {
// 				assert.NoError(t, err)
// 			},
// 			validate: func(t *testing.T, i *user.Item) {
// 				assert.NotNil(t, i)
// 				assert.Equal(t, uint(1), i.ID)
// 				assert.Equal(t, "testUser", i.Username)
// 			},
// 		},
// 		{
// 			name: "error_user_notfound",
// 			input: user.UpdateInput{
// 				Email:    testutil.PTRHelper("test123@gmail.com"),
// 				Password: testutil.PTRHelper("J@rovat1234"),
// 			},
// 			setup: func(ts *TestSuite) {
// 				m.On("GetByID", ts.Ctx, userID).Return(nil, apperror.ErrNotFound)
// 			},
// 			assertErr: func(t *testing.T, err error) {
// 				assert.Error(t, err)
// 				assert.ErrorIs(t, err, apperror.ErrNotFound)
// 			},
// 		},
// 		{
// 			name: "error_email_alreadyexist",
// 			input: user.UpdateInput{
// 				Email:    testutil.PTRHelper("conflict@mail.com"),
// 				Password: testutil.PTRHelper("J@rovat1234"),
// 			},
// 			setup: func(ts *TestSuite) {
// 				user := fixtures.ValidUser()
// 				m.On("GetByID", ts.Ctx, userID).Return(user, nil).Once()
// 				user.ID = 2
// 				user.Email = "conflict@mail.com"
// 				m.On("GetByEmail", ts.Ctx, "conflict@mail.com").Return(user, nil).Once()
// 			},
// 			assertErr: func(t *testing.T, err error) {
// 				assert.Error(t, err)
// 				assert.ErrorIs(t, err, apperror.ErrConflict)
// 			},
// 		},
// 		{
// 			name: "error_updated_fail",
// 			input: user.UpdateInput{
// 				Email: testutil.PTRHelper("conflict@mail.com"),
// 			},
// 			setup: func(ts *TestSuite) {
// 				user := fixtures.ValidUser()
// 				m.On("GetByID", ts.Ctx, userID).Return(user, nil).Once()
// 				m.On("GetByEmail", ts.Ctx, "conflict@mail.com").Return(nil, apperror.ErrNotFound).Once()
// 				m.On("Update", ts.Ctx, ts.CtxOfType("*domain.User")).Return(apperror.ErrInternalServer).Once()

// 			},
// 			assertErr: func(t *testing.T, err error) {
// 				assert.Error(t, err)
// 				assert.ErrorIs(t, err, apperror.ErrInternalServer)
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			mockRepo.ExpectedCalls = nil
// 			mockRepo.Calls = nil

// 			test.setup(mockRepo)

// 			user, err := newSV.UpdateUserProfile(context.Background(), uint(userID), test.input)
// 			if test.validate != nil {
// 				test.validate(t, user)
// 			}

// 			if test.assertErr != nil {
// 				test.assertErr(t, err)
// 			}

// 			mockRepo.AssertExpectations(t)
// 		})
// 	}
// }

// List test
// func TestUserService_List(t *testing.T) {
// 	mockRepo := mocks.NewMockUserRepository()
// 	newUC := user.NewUserUseCase(mockRepo)

// 	tests := []struct {
// 		name      string
// 		q         user.ListQuery
// 		setup     func(ts *TestSuite)
// 		assertErr func(*testing.T, error)
// 		validate  func(*testing.T, *user.ListOutput)
// 	}{
// 		{
// 			name: "success two items",
// 			q:    user.ListQuery{Limit: 10, Offset: 0},
// 			setup: func(ts *TestSuite) {
// 				u1 := fixtures.ValidUser()
// 				u2 := fixtures.ValidUser()
// 				u2.ID = 2
// 				u2.Email = "second@mail.com"

// 				m.On("List", ts.Ctx, mock.MatchedBy(func(q userrepo.ListQuery) bool {
// 					return q.Limit == 10 && q.Offset == 0 && q.Search == ""
// 				})).
// 					Return([]*domain.User{u1, u2}, int64(2), nil)
// 			},
// 			validate: func(t *testing.T, out *user.ListOutput) {
// 				require.NotNil(t, out)
// 				assert.Len(t, out.Items, 2)
// 				assert.Equal(t, int64(2), out.Total)
// 			},
// 		},
// 		{
// 			name: "apply default limit when zero",
// 			q:    user.ListQuery{Limit: 0, Offset: 5},
// 			setup: func(ts *TestSuite) {
// 				m.On("List", ts.Ctx, mock.MatchedBy(func(q userrepo.ListQuery) bool {
// 					return q.Limit == 10 && q.Offset == 5
// 				})).
// 					Return([]*domain.User{}, int64(0), nil)
// 			},
// 			validate: func(t *testing.T, out *user.ListOutput) {
// 				require.NotNil(t, out)
// 				assert.Len(t, out.Items, 0)
// 			},
// 		},
// 		{
// 			name: "cap limit when too large",
// 			q:    user.ListQuery{Limit: 500, Offset: 0},
// 			setup: func(ts *TestSuite) {
// 				m.On("List", ts.Ctx, mock.MatchedBy(func(q userrepo.ListQuery) bool {
// 					return q.Limit == 100 // capped

// 				})).
// 					Return([]*domain.User{}, int64(0), nil)
// 			},
// 			validate: func(t *testing.T, out *user.ListOutput) {
// 				require.NotNil(t, out)
// 				assert.Len(t, out.Items, 0)
// 			},
// 		},
// 		{
// 			name: "with search filter",
// 			q:    user.ListQuery{Search: "testUser"},
// 			setup: func(ts *TestSuite) {
// 				m.On("List", ts.Ctx, mock.MatchedBy(func(q userrepo.ListQuery) bool {
// 					return q.Search == "testUser"
// 				})).
// 					Return([]*domain.User{}, int64(0), nil)
// 			},
// 			validate: func(t *testing.T, out *user.ListOutput) {
// 				require.NotNil(t, out)
// 				assert.Len(t, out.Items, 0)
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			if test.setup != nil {
// 				test.setup(mockRepo)
// 			}

// 			out, err := newUC.List(context.Background(), test.q)

// 			if test.assertErr != nil {
// 				test.assertErr(t, err)
// 			} else {
// 				require.NoError(t, err)
// 			}
// 			if test.validate != nil {
// 				test.validate(t, out)
// 			}

// 			mockRepo.AssertExpectations(t)
// 		})
// 	}
// }
