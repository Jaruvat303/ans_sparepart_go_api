package category_test

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/category"
	mocks "ans-spareparts-api/internal/mock"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/testutil/fixtures"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ServiceTestSuite struct {
	Service      category.Service
	MockCategory *mocks.CategoryRepository
	Ctx          context.Context
}

func newServiceTestSuite() *ServiceTestSuite {
	return &ServiceTestSuite{}
}

func (ts *ServiceTestSuite) SetupTest(t *testing.T) {
	// สร้าง Mock
	ts.MockCategory = mocks.NewMockCategoryRepository()
	ts.Ctx = context.Background()

	ts.Service = category.NewService(ts.MockCategory)

	// ตั้งค่า Teardown: จะถูกเรียกเมื่อ t.Run หรือ test func จบ
	t.Cleanup(func() {
		// ตรวจสอบว่า AssetExpectation ถูกเรียกใช้แต่ละ test
		ts.MockCategory.AssertExpectations(t)
	})
}

func TestCategoryService_Create(t *testing.T) {
	mockCategory := fixtures.ValidCategory()
	mockRequest := category.CategoryRequest{
		Name: "Wheel",
	}

	tests := []struct {
		name      string
		input     category.CategoryRequest
		setup     func(*ServiceTestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *category.Item)
	}{
		{
			name:  "create_success",
			input: mockRequest,
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByName", ts.Ctx, "Wheel").Return(nil, apperror.ErrNotFound).Once()
				ts.MockCategory.On("Create", ts.Ctx, mock.MatchedBy(func(cat *domain.Category) bool {
					return cat.ID == 0 && cat.Name == "Wheel"
				})).Return(nil).Once().Run(func(args mock.Arguments) {
					// จำลองการเปลี่ยนแปลง Object หลังจากบันทึกข้อมูลลง DB
					createCat := args.Get(1).(*domain.Category)
					createCat.ID = 1
				})
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *category.Item) {
				assert.NotNil(t, i)
				assert.Equal(t, mockCategory.ID, i.ID)
				assert.Equal(t, mockCategory.Name, i.Name)
			},
		},
		{
			name:  "create_fail_invalid_name",
			input: mockRequest,
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByName", ts.Ctx, "Wheel").Return(mockCategory, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, apperror.ErrConflict)
			},
			validate: func(t *testing.T, i *category.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name:  "create_fail_db_erorr",
			input: mockRequest,
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByName", ts.Ctx, "Wheel").Return(nil, apperror.ErrNotFound)
				ts.MockCategory.On("Create", ts.Ctx, mock.AnythingOfType("*domain.Category")).Return(apperror.ErrInternalServer)
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, apperror.ErrInternalServer)
			},
			validate: func(t *testing.T, i *category.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Setup ServiceTestsuite
			ts := newServiceTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			item, err := ts.Service.CreateCategory(ts.Ctx, test.input)

			test.assertErr(t, err)

			test.validate(t, item)

		})
	}
}

func TestCategoryService_Update(t *testing.T) {
	mockCategory := fixtures.ValidCategory()
	mockRequest := category.CategoryRequest{
		Name: "Wheel",
	}
	tests := []struct {
		name      string
		catID     uint
		input     category.CategoryRequest
		setup     func(ts *ServiceTestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *category.Item)
	}{
		{
			name:  "update_success",
			catID: uint(1),
			input: mockRequest,
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByName", ts.Ctx, "Wheel").Return(nil, apperror.ErrNotFound).Once()
				ts.MockCategory.On("GetByID", ts.Ctx, uint(1)).Return(mockCategory, nil).Once()
				ts.MockCategory.On("Update", ts.Ctx, mock.MatchedBy(func(cat *domain.Category) bool {
					return cat.ID == uint(1) && cat.Name == "Wheel"
				})).Return(nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *category.Item) {
				assert.NotNil(t, i)
				assert.Equal(t, mockCategory.Name, i.Name)
				assert.Equal(t, mockCategory.ID, i.ID)
			},
		},
		{
			name:  "update_error_category_name_already_exists",
			catID: uint(1),
			input: mockRequest,
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByName", ts.Ctx, "Wheel").Return(mockCategory, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.ErrorIs(t, apperror.ErrConflict, err)
			},
			validate: func(t *testing.T, i *category.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name:  "update_error_category_not_found",
			catID: uint(999),
			input: mockRequest,
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByName", ts.Ctx, "Wheel").Return(nil, apperror.ErrNotFound)
				ts.MockCategory.On("GetByID", ts.Ctx, uint(999)).Return(nil, apperror.ErrNotFound)
			},
			assertErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.ErrorIs(t, apperror.ErrNotFound, err)
			},
			validate: func(t *testing.T, i *category.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name:  "update_db_error",
			catID: uint(1),
			input: mockRequest,
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByName", ts.Ctx, "Wheel").Return(nil, apperror.ErrNotFound).Once()
				ts.MockCategory.On("GetByID", ts.Ctx, uint(1)).Return(mockCategory, nil).Once()
				ts.MockCategory.On("Update", ts.Ctx, mock.MatchedBy(func(cat *domain.Category) bool {
					return cat.ID == uint(1) && cat.Name == "Wheel"
				})).Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *category.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := newServiceTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			item, err := ts.Service.UpdateCategory(ts.Ctx, test.catID, test.input)

			test.assertErr(t, err)
			test.validate(t, item)
		})
	}
}

func TestCategoryService_Delete(t *testing.T) {
	validCategory := fixtures.ValidCategory()

	tests := []struct {
		name      string
		catID     uint
		setup     func(ts *ServiceTestSuite)
		assertErr func(*testing.T, error)
	}{
		{
			name:  "delete_successfull",
			catID: uint(1),
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil)
				ts.MockCategory.On("Delete", ts.Ctx, uint(1)).Return(nil)
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "delete_error_category_not_found",
			catID: uint(999),
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByID", ts.Ctx, uint(999)).Return(nil, apperror.ErrNotFound)
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, apperror.ErrNotFound, err)

			},
		}, {
			name:  "delete_dberror",
			catID: uint(1),
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil).Once()
				ts.MockCategory.On("Delete", ts.Ctx, uint(1)).Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := newServiceTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			err := ts.Service.DeleteCategory(ts.Ctx, test.catID)

			test.assertErr(t, err)
		})
	}
}

func TestCategoryService_GetByID(t *testing.T) {
	validCategory := fixtures.ValidCategory()
	tests := []struct {
		name      string
		catID     uint
		setup     func(*ServiceTestSuite)
		assertErr func(*testing.T, error)
		Validate  func(*testing.T, *category.Item)
	}{
		{
			name:  "getbyid_successfull",
			catID: uint(1),
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil)
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			Validate: func(t *testing.T, i *category.Item) {
				assert.NotNil(t, i)
				assert.Equal(t, validCategory.ID, i.ID)
				assert.Equal(t, validCategory.Name, i.Name)
			},
		},
		{
			name:  "getbyid_category_not_found",
			catID: uint(99),
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByID", ts.Ctx, uint(99)).Return(nil, apperror.ErrNotFound)
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, apperror.ErrNotFound, err)
			},
			Validate: func(t *testing.T, i *category.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name:  "getbyid_category_internal_error",
			catID: uint(1),
			setup: func(ts *ServiceTestSuite) {
				ts.MockCategory.On("GetByID", ts.Ctx, uint(1)).Return(nil, apperror.ErrInternalServer)
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			Validate: func(t *testing.T, i *category.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := newServiceTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			item, err := ts.Service.GetCategoryByID(ts.Ctx, test.catID)

			test.assertErr(t, err)
			test.Validate(t, item)
		})
	}
}

func TestCategoryService_List(t *testing.T) {
	validCategory := fixtures.ValidListCategory()

	tests := []struct {
		name     string
		input    category.ListQuery
		setup    func(ts *ServiceTestSuite)
		asserErr func(*testing.T, error)
		validate func(*testing.T, *category.ListOutput)
	}{
		{
			name: "search_by_name_specific_pagination_success",
			input: category.ListQuery{
				Search: "Wheel",
			},
			setup: func(ts *ServiceTestSuite) {
				expectedQuery := category.ListQuery{
					Search: "Wheel",
					Limit:  10,
					Offset: 0,
				}
				ts.MockCategory.On("List", ts.Ctx, expectedQuery).Return(validCategory, int64(len(validCategory)), nil).Once()
			},
			asserErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, clr *category.ListOutput) {
				assert.NotNil(t, clr)
				assert.Equal(t, validCategory[0].ID, clr.Items[0].ID)
				assert.Equal(t, validCategory[0].Name, clr.Items[0].Name)
				assert.Equal(t, int64(len(validCategory)), clr.Total)
			},
		},
		{
			name: "search_dberror",
			input: category.ListQuery{
				Limit: 20,
			},
			setup: func(ts *ServiceTestSuite) {
				expectedQuery := category.ListQuery{
					Search: "",
					Limit:  20,
					Offset: 0,
					Sort:   "",
				}

				ts.MockCategory.On("List", ts.Ctx, expectedQuery).Return(nil, int64(0), apperror.ErrInternalServer)
			},
			asserErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, lo *category.ListOutput) {
				assert.Nil(t, lo)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := newServiceTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			item, err := ts.Service.List(ts.Ctx, test.input)

			test.asserErr(t, err)
			test.validate(t, item)
		})
	}
}
