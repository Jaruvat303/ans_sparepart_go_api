package product_test

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/category"
	"ans-spareparts-api/internal/features/inventory"
	"ans-spareparts-api/internal/features/product"
	mocks "ans-spareparts-api/internal/mock"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/testutil"
	"ans-spareparts-api/pkg/testutil/fixtures"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestSuite struct: เก้บตัวแปรที่ใช้ร่วมกัน
type TestSuite struct {
	Service           product.Service
	MockProductRepo   *mocks.ProductRepository
	MockCategoryRepo  *mocks.CategoryRepository
	MockInventoryRepo *mocks.InventoryRepository
	Ctx               context.Context
}

func NewTestSuite() *TestSuite {
	return &TestSuite{}
}

// SetupTest คือ Helper function ที่สร้าง Instance object ที่ใช้ร่วมกันแต่ละ function
func (ts *TestSuite) SetupTest(t *testing.T) {
	// สร้าง Mock Instance
	ts.MockProductRepo = mocks.NewMockProductRepository()
	ts.MockCategoryRepo = mocks.NewMockCategoryRepository()
	ts.MockInventoryRepo = mocks.NewMockInventoryRepository()
	ts.Ctx = context.Background()

	// สร้าง service Instance โดยใช้ข้่อมูล mock
	ts.Service = product.NewService(
		ts.MockProductRepo,
		ts.MockCategoryRepo,
		ts.MockInventoryRepo,
	)

	// ตั้งค่า Teardown: จะถูกเรียกเมื่อ t.Run หรือ test func จบ
	t.Cleanup(func() {
		// ตรวจสอบว่า AssetExpectation ถูกเรียกใช้แต่ละ test
		ts.MockProductRepo.AssertExpectations(t)
		ts.MockCategoryRepo.AssertExpectations(t)
		ts.MockInventoryRepo.AssertExpectations(t)
	})
}

func TestProductService_GetProductDetail(t *testing.T) {

	// ข้อมูลสินค้าที่ใช้ในกรณีสำเร็จ (Success Case)
	validProduct := fixtures.ValidProduct()
	validCategory := fixtures.ValidCategory()
	validInventory := fixtures.ValidInventory()

	// Response ที่คาดหวัง (
	expectedItem := &product.Item{
		ID:         validProduct.ID,
		Name:       validProduct.Name,
		CategoryID: validCategory.ID,
		Category: category.CategoryResponse{
			ID:   validCategory.ID,
			Name: validCategory.Name,
		},
		Inventory: inventory.InventoryResponse{
			ID:        validInventory.ID,
			ProductID: validInventory.ProductID,
			Quantity:  validInventory.Quantity,
		},
	}

	tests := []struct {
		name      string
		inputID   uint
		setup     func(ts *TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *product.Item)
	}{
		{
			name:    "Success_All_Data_Found",
			inputID: 1,
			setup: func(ts *TestSuite) {
				// Mock การเรียก Repository ทั้ง 3 ครั้ง (สำเร็จทั้งหมด)
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil).Once()
				ts.MockInventoryRepo.On("GetByProductID", ts.Ctx, uint(1)).Return(validInventory, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, item *product.Item) {
				assert.NotNil(t, item)
				assert.Equal(t, expectedItem.ID, item.ID)
				assert.Equal(t, expectedItem.Category.Name, item.Category.Name)
				assert.Equal(t, expectedItem.Inventory.Quantity, item.Inventory.Quantity)
			},
		},
		{
			name:    "Error_ProductRepo_Failed",
			inputID: 2,
			setup: func(ts *TestSuite) {
				// Mock: GetByID ของ Product Repo คืน Error
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(2)).Return(nil, apperror.ErrInternalServer).Once()
				// Repository อื่นๆ ไม่ถูกเรียก
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, item *product.Item) {
				assert.Nil(t, item)
			},
		},
		{
			name:    "Error_CategoryRepo_Failed",
			inputID: 3,
			setup: func(ts *TestSuite) {

				expectedErr := errors.New("category connection failed")
				// Mock: Product Repo สำเร็จ
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(3)).Return(validProduct, nil).Once()
				// Mock: Category Repo คืน Error
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, validProduct.CategoryID).Return(nil, expectedErr).Once()
				// Repository อื่นๆ ไม่ถูกเรียก
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "category connection failed")
			},
			validate: func(t *testing.T, item *product.Item) {
				assert.Nil(t, item)
			},
		},
		{
			name:    "Error_InventoryRepo_Failed",
			inputID: 4,
			setup: func(ts *TestSuite) {
				expectedErr := errors.New("inventory service unreachable")
				// Mock: Product Repo สำเร็จ
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(4)).Return(validProduct, nil).Once()
				// Mock: Category Repo สำเร็จ
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, validProduct.CategoryID).Return(validCategory, nil).Once()
				// Mock: Inventory Repo คืน Error
				ts.MockInventoryRepo.On("GetByProductID", ts.Ctx, uint(4)).Return(nil, expectedErr).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "inventory service unreachable")
			},
			validate: func(t *testing.T, item *product.Item) {
				assert.Nil(t, item)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// สร้าง Test Suite Instance ใหม่ของแต่ละ t.Run
			testsuite := NewTestSuite()
			testsuite.SetupTest(t) // SetupTest ที่สร้าง mock/service ใหม่และตั้งค่า t.cleanup

			// รัน Setup ของ Test Case นั้น
			test.setup(testsuite)

			// Execute Usecase โดยใช้ service ของ TestSuite
			item, err := testsuite.Service.GetProductDetail(testsuite.Ctx, test.inputID)

			// Assertions and Validate
			test.assertErr(t, err)
			test.validate(t, item)

			// t.Cleanup จะเรียก mock.AssertExpectatio(t) อัตโนมัติเมื่อจบ t.Run
		})
	}
}

func TestProductService_CreateProduct(t *testing.T) {
	// ข้อมูลสำหรับ Success Case
	validProduct := fixtures.ValidProduct()
	validCategory := fixtures.ValidCategory()
	validInventory := fixtures.ValidInventory()

	// expected Item
	expectedItem := &product.Item{
		ID:         validProduct.ID,
		Name:       validProduct.Name,
		CategoryID: validCategory.ID,
		Category: category.CategoryResponse{
			ID:   validCategory.ID,
			Name: validCategory.Name,
		},
		Inventory: inventory.InventoryResponse{
			ID:        validInventory.ID,
			ProductID: validInventory.ProductID,
			Quantity:  validInventory.Quantity,
		},
	}

	tests := []struct {
		name      string
		input     product.CreateInput
		setup     func(*TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *product.Item)
	}{
		{
			name: "Success_NewProduct_Created",
			input: product.CreateInput{
				Name:        "Test",
				Description: "Description",
				SKU:         "SKU",
				Price:       1,
				CategoryID:  1,
			},
			setup: func(ts *TestSuite) {
				// ตรวจสอบค่า SKU
				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "SKU").Return(nil, apperror.ErrNotFound).Once()
				// ตรวจสอบค่า CategoryID
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil).Once()
				// สร้างข้อมูล Product
				ts.MockProductRepo.On("Create", ts.Ctx, mock.MatchedBy(func(p *domain.Product) bool {
					// ตรวจสอบค่าก่อนบันทึก
					assert.Equal(t, "Test", p.Name)
					assert.Equal(t, "SKU", p.SKU)

					// จำลองการบันทึกค่า
					p.ID = 1
					return true
				})).Return(nil).Once()
				// สร้าง Inventory
				ts.MockInventoryRepo.On("Create", ts.Ctx, mock.MatchedBy(func(i *domain.Inventory) bool {
					i.ID = 1
					return true
				})).Return(validInventory, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Equal(t, expectedItem.ID, i.ID)
				assert.Equal(t, expectedItem.Name, i.Name)
				assert.Equal(t, expectedItem.Category.ID, i.Category.ID)
				assert.Equal(t, expectedItem.Inventory.ID, i.Inventory.ID)
			},
		},
		{
			name: "Error_InputInvalid_EmptyName",
			input: product.CreateInput{
				Name:        "",
				Description: "",
				SKU:         "SKU",
				Price:       1,
				CategoryID:  1,
			},
			setup: func(ts *TestSuite) {
				// mock function จะยังไม่ทำงานเพราะตรวจค่า Input ไม่ผ่าน
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInvalidInput, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_SKUNormalization",
			input: product.CreateInput{
				Name:        "Test",
				Description: "desc",
				SKU:         "*Sku",
				CategoryID:  1,
				Price:       1,
			},
			setup: func(ts *TestSuite) {
				// ไม่มี mocks ใดทำงาน error ตอน nornalization
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInvalidSKU, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_SKU_AlreadyExists",
			input: product.CreateInput{
				Name:        "Test",
				Description: "Desc",
				SKU:         "TestSku",
				Price:       1,
				CategoryID:  1,
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "TestSku").Return(validProduct, nil).Once()
				// mock อื่นๆจะไม่ทำงานจะ return Error ก่อน
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrConflict, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_SKUCheck_DBError",
			input: product.CreateInput{
				Name:        "Test",
				Description: "desc",
				SKU:         "SKU",
				Price:       1,
				CategoryID:  1,
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "SKU").Return(nil, apperror.ErrInternalServer)

			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_Category_NotFound",
			input: product.CreateInput{
				Name:        "Test",
				Description: "desc",
				SKU:         "SKU",
				Price:       1,
				CategoryID:  999,
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "SKU").Return(nil, apperror.ErrNotFound)
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(999)).Return(nil, apperror.ErrNotFound)
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrNotFound, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_ProductDB_CreateError",
			input: product.CreateInput{
				Name:        "Test",
				Description: "desc",
				SKU:         "SKU",
				Price:       1,
				CategoryID:  1,
			},
			setup: func(ts *TestSuite) {

				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "SKU").Return(nil, apperror.ErrNotFound).Once()
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil).Once()
				ts.MockProductRepo.On("Create", ts.Ctx, mock.MatchedBy(func(p *domain.Product) bool {
					p.ID = 1
					return true
				})).Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_InventoryDB_CreateError",
			input: product.CreateInput{
				Name:        "Test",
				Description: "desc",
				SKU:         "SKU",
				Price:       1,
				CategoryID:  1,
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "SKU").Return(nil, apperror.ErrNotFound).Once()
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil).Once()
				ts.MockProductRepo.On("Create", ts.Ctx, mock.MatchedBy(func(p *domain.Product) bool {
					p.ID = 1
					return true
				})).Return(nil).Once()
				ts.MockInventoryRepo.On("Create", ts.Ctx, mock.MatchedBy(func(i *domain.Inventory) bool {
					i.ProductID = uint(1)
					i.Quantity = 1
					return true
				})).Return(nil, apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTest(t)

			test.setup(ts)

			item, err := ts.Service.CreateProduct(ts.Ctx, test.input)

			test.assertErr(t, err)
			test.validate(t, item)
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	validProduct := fixtures.ValidProductLite()
	validCategory := fixtures.ValidCategory()
	validInventory := fixtures.ValidInventory()

	expectedItem := &product.Item{
		ID:          validProduct.ID,
		Name:        validProduct.Name,
		Description: validProduct.Description,
		SKU:         validProduct.SKU,
		Price:       validProduct.Price,
		CategoryID:  validProduct.CategoryID,
		Category: category.CategoryResponse{
			ID:   validCategory.ID,
			Name: validCategory.Name,
		},
		Inventory: inventory.InventoryResponse{
			ID:        validInventory.ID,
			ProductID: validInventory.ProductID,
			Quantity:  validInventory.Quantity,
		},
	}

	tests := []struct {
		name      string
		ID        uint
		input     product.UpdateInput
		setup     func(ts *TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *product.Item)
	}{
		{
			name: "Success_Update_AllFields",
			ID:   1,
			input: product.UpdateInput{
				Name:        testutil.PTRHelper("Test"),
				Description: testutil.PTRHelper("Description"),
				SKU:         testutil.PTRHelper("SKU"),
				Price:       testutil.PTRHelper(1.0),
				CategoryID:  testutil.PTRHelper(uint(1)),
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "SKU").Return(nil, apperror.ErrNotFound).Once()
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil).Once()
				ts.MockProductRepo.On("Update", ts.Ctx, mock.MatchedBy(func(p *domain.Product) bool {
					assert.Equal(t, uint(1), p.ID)
					assert.Equal(t, "Test", p.Name)
					assert.Equal(t, "Description", p.Description)
					assert.Equal(t, "SKU", p.SKU)
					assert.Equal(t, 1.0, p.Price)
					assert.Equal(t, uint(1), p.CategoryID)
					return true
				})).Return(nil).Once()
				ts.MockInventoryRepo.On("GetByProductID", ts.Ctx, uint(1)).Return(validInventory, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.NotNil(t, i)
				assert.Equal(t, expectedItem.ID, i.ID)
				assert.Equal(t, expectedItem.Name, i.Name)
				assert.Equal(t, expectedItem.Description, i.Description)
				assert.Equal(t, expectedItem.SKU, i.SKU)
				assert.Equal(t, expectedItem.Price, i.Price)
				assert.Equal(t, expectedItem.Category, i.Category)
			},
		},
		{
			name: "Success_Update_Partial_Name_Price",
			ID:   uint(1),
			input: product.UpdateInput{
				Name:  testutil.PTRHelper("New Name"),
				Price: testutil.PTRHelper(99.99),
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockProductRepo.On("Update", ts.Ctx, mock.MatchedBy(func(p *domain.Product) bool {
					return p.ID == 1 && p.Name == "New Name" && p.Price == 99.99
				})).Return(nil).Once()
				ts.MockInventoryRepo.On("GetByProductID", ts.Ctx, uint(1)).Return(validInventory, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.NotNil(t, i)
				assert.Equal(t, uint(1), i.ID)
				assert.Equal(t, "New Name", i.Name)
				assert.Equal(t, 99.99, i.Price)
			},
		},
		{
			name: "Error_Product_NotFound",
			ID:   uint(999),
			input: product.UpdateInput{
				Name: testutil.PTRHelper("New Name"),
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(999)).Return(nil, apperror.ErrNotFound).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrNotFound, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_InputValidation_InvalidPrice",
			ID:   uint(1),
			input: product.UpdateInput{
				Price: testutil.PTRHelper(-99.9),
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()

			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, apperror.ErrInvalidInput, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_Categoty_Notfound",
			ID:   uint(1),
			input: product.UpdateInput{
				CategoryID: testutil.PTRHelper(uint(99)),
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(99)).Return(nil, apperror.ErrNotFound).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrNotFound, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_Categoty_DBError",
			ID:   uint(1),
			input: product.UpdateInput{
				CategoryID: testutil.PTRHelper(uint(99)),
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(99)).Return(nil, apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Error_Update_DBError",
			ID:   1,
			input: product.UpdateInput{
				Name:        testutil.PTRHelper("Test"),
				Description: testutil.PTRHelper("Description"),
				SKU:         testutil.PTRHelper("SKU"),
				Price:       testutil.PTRHelper(1.0),
				CategoryID:  testutil.PTRHelper(uint(1)),
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "SKU").Return(nil, apperror.ErrNotFound).Once()
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil).Once()
				ts.MockProductRepo.On("Update", ts.Ctx, mock.MatchedBy(func(p *domain.Product) bool {
					assert.Equal(t, uint(1), p.ID)
					assert.Equal(t, "Test", p.Name)
					assert.Equal(t, "Description", p.Description)
					assert.Equal(t, "SKU", p.SKU)
					assert.Equal(t, 1.0, p.Price)
					assert.Equal(t, uint(1), p.CategoryID)
					return true
				})).Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "Success_Update_AllFields",
			ID:   1,
			input: product.UpdateInput{
				Name:        testutil.PTRHelper("Test"),
				Description: testutil.PTRHelper("Description"),
				SKU:         testutil.PTRHelper("SKU"),
				Price:       testutil.PTRHelper(1.0),
				CategoryID:  testutil.PTRHelper(uint(1)),
			},
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockProductRepo.On("GetBySKU", ts.Ctx, "SKU").Return(nil, apperror.ErrNotFound).Once()
				ts.MockCategoryRepo.On("GetByID", ts.Ctx, uint(1)).Return(validCategory, nil).Once()
				ts.MockProductRepo.On("Update", ts.Ctx, mock.MatchedBy(func(p *domain.Product) bool {
					assert.Equal(t, uint(1), p.ID)
					assert.Equal(t, "Test", p.Name)
					assert.Equal(t, "Description", p.Description)
					assert.Equal(t, "SKU", p.SKU)
					assert.Equal(t, 1.0, p.Price)
					assert.Equal(t, uint(1), p.CategoryID)
					return true
				})).Return(nil).Once()
				ts.MockInventoryRepo.On("GetByProductID", ts.Ctx, uint(1)).Return(nil, apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *product.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			item, err := ts.Service.UpdateProduct(ts.Ctx, test.ID, test.input)

			test.assertErr(t, err)
			test.validate(t, item)
		})
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	validProduct := fixtures.ValidListProduct()
	productID := uint(1)

	tests := []struct {
		name      string
		input     uint
		setup     func(ts *TestSuite)
		assertErr func(*testing.T, error)
	}{
		{
			name:  "Success_Product_Deleted",
			input: productID,
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockProductRepo.On("Delete", ts.Ctx, uint(1)).Return(nil).Once()
				ts.MockInventoryRepo.On("Delete", ts.Ctx, uint(1)).Return(nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:  "Error_Product_NotFound",
			input: uint(999),
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(999)).Return(nil, apperror.ErrNotFound).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrNotFound, err)
			},
		},
		{
			name:  "Error_Delete_Product_DBError",
			input: productID,
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockProductRepo.On("Delete", ts.Ctx, uint(1)).Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
		},
		{
			name:  "Error_Delete_Inventory_DBError",
			input: productID,
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("GetByID", ts.Ctx, uint(1)).Return(validProduct, nil).Once()
				ts.MockProductRepo.On("Delete", ts.Ctx, uint(1)).Return(nil).Once()
				ts.MockInventoryRepo.On("Delete", ts.Ctx, uint(1)).Return(apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			err := ts.Service.DeleteProduct(ts.Ctx, test.input)

			test.assertErr(t, err)
		})
	}
}

func TestProductService_List(t *testing.T) {
	validateProduct := fixtures.ValidListProduct()
	inputQuery := product.ListQuery{
		Search: "product",
		Limit:  10,
		Offset: 20,
		Sort:   "name desc",
	}

	tests := []struct {
		name      string
		input     product.ListQuery
		setup     func(ts *TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *product.ListOutput)
	}{
		{
			name:  "Success_Retrieves_MultipleProducts",
			input: inputQuery,
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("List", ts.Ctx, mock.MatchedBy(func(q product.ListQuery) bool {
					assert.Equal(t, "product", q.Search)
					assert.Equal(t, 10, q.Limit)
					assert.Equal(t, 20, q.Offset)
					assert.Equal(t, "name desc", q.Sort)
					return true
				})).Return(validateProduct, int64(2), nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, lo *product.ListOutput) {
				assert.NotNil(t, lo)
				assert.Equal(t, int64(2), lo.Total)
				assert.Equal(t, validateProduct[0].ID, lo.Items[0].ID)

			},
		},
		{
			name:  "Success_Retrieves_NoProducts",
			input: inputQuery,
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("List", ts.Ctx, mock.MatchedBy(func(q product.ListQuery) bool {
					assert.Equal(t, "product", q.Search)
					assert.Equal(t, 10, q.Limit)
					assert.Equal(t, 20, q.Offset)
					assert.Equal(t, "name desc", q.Sort)
					return true
				})).Return([]*domain.Product{}, int64(0), nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, lo *product.ListOutput) {
				assert.NotNil(t, lo)

			},
		},
		{
			name:  "Error_Retrieves_DBError",
			input: inputQuery,
			setup: func(ts *TestSuite) {
				ts.MockProductRepo.On("List", ts.Ctx, mock.MatchedBy(func(q product.ListQuery) bool {
					assert.Equal(t, "product", q.Search)
					assert.Equal(t, 10, q.Limit)
					assert.Equal(t, 20, q.Offset)
					assert.Equal(t, "name desc", q.Sort)
					return true
				})).Return(nil, int64(0), apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, lo *product.ListOutput) {
				assert.Nil(t, lo)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			output, err := ts.Service.List(ts.Ctx, test.input)

			test.assertErr(t, err)
			test.validate(t, output)
		})
	}
}
