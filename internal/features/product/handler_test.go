package product_test

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/category"
	"ans-spareparts-api/internal/features/inventory"
	"ans-spareparts-api/internal/features/product"
	"ans-spareparts-api/internal/infra/jwtx"
	mocks "ans-spareparts-api/internal/mock"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/response"
	"ans-spareparts-api/pkg/testutil"
	"ans-spareparts-api/pkg/testutil/fixtures"
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type HandlerTestSuite struct {
	App         *fiber.App
	MockService *mocks.ProductService
	Handler     *product.Handler
}

func NewHandlerTestSuite() *HandlerTestSuite {
	return &HandlerTestSuite{}
}

func (ts *HandlerTestSuite) SetUpHandlerTestSuite(t *testing.T) {
	ts.MockService = mocks.NewProductService()
	ts.Handler = product.NewHandler(ts.MockService)
	ts.App = fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body: "+err.Error())
		},
	})

	t.Cleanup(func() {
		ts.MockService.AssertExpectations(t)
	})
}

var mockProduct = fixtures.ValidProduct()
var mockCategory = fixtures.ValidCategory()
var mockInventory = fixtures.ValidInventory()

var mockItem = &product.Item{
	ID:          mockProduct.ID,
	Name:        mockProduct.Name,
	Description: mockProduct.Description,
	Price:       mockProduct.Price,
	SKU:         mockProduct.SKU,
	CategoryID:  mockProduct.CategoryID,
	Category: category.CategoryResponse{
		ID:   mockCategory.ID,
		Name: mockCategory.Name,
	},
	Inventory: inventory.InventoryResponse{
		ID:        mockInventory.ID,
		ProductID: mockInventory.ID,
		Quantity:  mockInventory.Quantity,
	},
}
var mockResponse = &product.ProductDetailResponse{
	ID:          mockProduct.ID,
	Name:        mockProduct.Name,
	Description: mockProduct.Description,
	Price:       mockProduct.Price,
	SKU:         mockProduct.SKU,
	CategoryID:  mockProduct.CategoryID,
	Category: category.CategoryResponse{
		ID:   mockCategory.ID,
		Name: mockCategory.Name,
	},
	Inventory: inventory.InventoryResponse{
		ID:        mockInventory.ID,
		ProductID: mockInventory.ProductID,
		Quantity:  mockInventory.Quantity,
	},
}

func TestProductHandler_CreateProduct(t *testing.T) {

	mockRequest := product.CreateProductRequest{
		Name:        mockProduct.Name,
		Description: mockProduct.Description,
		Price:       mockProduct.Price,
		SKU:         mockProduct.SKU,
		CategoryID:  mockProduct.CategoryID,
	}
	mockInput := product.CreateInput{
		Name:        mockProduct.Name,
		Description: mockProduct.Description,
		Price:       mockProduct.Price,
		SKU:         mockProduct.SKU,
		CategoryID:  mockProduct.CategoryID,
	}

	tests := []struct {
		name               string
		userRole           string
		path               string
		requestBody        interface{}
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name:        "Success_Create_Product_With_Manager",
			userRole:    "manager",
			path:        "/products",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("CreateProduct", mock.Anything, mockInput).Return(mockItem, nil).Once()
			},
			expectedStatusCode: fiber.StatusCreated,
			expectedBody:       mockResponse,
		},
		{
			name:               "Error_Forbidden_With_Cashier",
			userRole:           "cashier",
			path:               "/products",
			requestBody:        mockRequest,
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusForbidden,
			expectedBody: fiber.Map{

				"code":    "FORBIDDEN",
				"message": "insufficient permission",
			},
		},
		{
			name:               "Error_BadRequest_Invalid_RequestBody",
			userRole:           "manager",
			path:               "/products",
			requestBody:        `{"name":"missing double quote}`,
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		},
		{
			name:        "Error_Conflict_Product_Already",
			userRole:    "manager",
			path:        "/products",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("CreateProduct", mock.Anything, mockInput).Return(nil, apperror.ErrConflict).Once()
			},
			expectedStatusCode: fiber.StatusConflict,
			expectedBody: fiber.Map{

				"code":    "CONFLICT",
				"message": "product alreadt exist",
			},
		},
		{
			name:        "Error_Category_NotFound",
			userRole:    "manager",
			path:        "/products",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("CreateProduct", mock.Anything, mockInput).Return(nil, apperror.ErrNotFound).Once()
			},
			expectedStatusCode: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "category not found",
			},
		},
		{
			name:        "Error_InternalServer",
			userRole:    "manager",
			path:        "/products",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("CreateProduct", mock.Anything, mockInput).Return(nil, apperror.ErrInternalServer).Once()
			},
			expectedStatusCode: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetUpHandlerTestSuite(t)

			// mock jwt cliams
			ts.App.Use(func(c *fiber.Ctx) error {
				if test.userRole != "" {
					c.Locals("user", &jwtx.Claims{UserID: 1, Username: "Test", Role: test.userRole})
				}
				return c.Next()
			})

			// create route
			ts.App.Post("/products", ts.Handler.CreateProduct)
			test.setup(ts)

			var body []byte
			if value, ok := test.requestBody.(string); ok {
				body = []byte(value)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			// Create request
			req := httptest.NewRequest(fiber.MethodPost, test.path, bytes.NewBuffer(body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Run test
			res, _ := ts.App.Test(req, -1)

			// status code check
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			// Change Data type to byte
			expectedBody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resBody))

		})
	}
}

func TestProductHandler_GetProductDetail(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		setup              func(*HandlerTestSuite)
		expectedStatasCode int
		expectedBody       interface{}
	}{
		{
			name: "Success_Get_Product_Detail",
			path: "/products/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetProductDetail", mock.Anything, uint(1)).Return(mockItem, nil).Once()
			},
			expectedStatasCode: fiber.StatusOK,
			expectedBody:       mockResponse,
		},
		{
			name:               "Error_Invalid_productID_Format",
			path:               "/products/one",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatasCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid productID",
			},
		},

		{
			name: "Error_Product_NotFound",
			path: "/products/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetProductDetail", mock.Anything, uint(1)).Return(nil, apperror.ErrNotFound).Once()
			},
			expectedStatasCode: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "product not found",
			},
		},
		{
			name: "Error_InternalServer",
			path: "/products/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetProductDetail", mock.Anything, uint(1)).Return(nil, apperror.ErrInternalServer).Once()
			},
			expectedStatasCode: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetUpHandlerTestSuite(t)

			// Create Route
			ts.App.Get("/products/:id", ts.Handler.GetProductDetail)
			test.setup(ts)

			// create request
			req := httptest.NewRequest(fiber.MethodGet, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// runtest
			res, _ := ts.App.Test(req, -1)

			// Status check
			assert.Equal(t, test.expectedStatasCode, res.StatusCode)

			// tranform data type to byte
			expectedBody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resBody))
		})
	}
}

func TestProductHandler_UpdateProduct(t *testing.T) {
	mockRequest := &product.UpdateProductRequest{
		Name:        testutil.PTRHelper(mockProduct.Name),
		Description: testutil.PTRHelper(mockProduct.Description),
		SKU:         testutil.PTRHelper(mockProduct.SKU),
		Price:       testutil.PTRHelper(mockProduct.Price),
		CategoryID:  testutil.PTRHelper(mockProduct.CategoryID),
	}
	mockInput := product.UpdateInput{
		Name:        mockRequest.Name,
		Description: mockRequest.Description,
		SKU:         mockRequest.SKU,
		Price:       mockRequest.Price,
		CategoryID:  mockRequest.CategoryID,
	}

	tests := []struct {
		name               string
		userRole           string
		path               string
		requestBody        interface{}
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name:        "Success_Update_AllField",
			userRole:    "manager",
			path:        "/products/1",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateProduct", mock.Anything, uint(1), mockInput).Return(mockItem, nil).Once()
			},
			expectedStatusCode: fiber.StatusOK,
			expectedBody:       mockResponse,
		},
		{
			name:     "Success_Update_FieldName",
			userRole: "manager",
			path:     "/products/1",
			requestBody: product.UpdateInput{
				Name: testutil.PTRHelper(mockProduct.Name),
			},
			setup: func(hts *HandlerTestSuite) {
				input := product.UpdateInput{
					Name: testutil.PTRHelper(mockProduct.Name),
				}
				hts.MockService.On("UpdateProduct", mock.Anything, uint(1), input).Return(mockItem, nil).Once()
			},
			expectedStatusCode: fiber.StatusOK,
			expectedBody:       mockResponse,
		},
		{
			name:               "Error_Forbridden_Update_With_Cashier",
			userRole:           "cashier",
			path:               "/products/1",
			requestBody:        mockRequest,
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusForbidden,
			expectedBody: fiber.Map{

				"code":    "FORBIDDEN",
				"message": "insufficient permission",
			},
		},
		{
			name:               "Error_BadRequest_InvalidProductID_Format",
			userRole:           "manager",
			path:               "/products/one",
			requestBody:        mockRequest,
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid productID",
			},
		},
		{
			name:        "Error_BadRequest_InvalidBody",
			userRole:    "manager",
			path:        "/products/1",
			requestBody: `{"name":"missing double quote}`,
			setup: func(hts *HandlerTestSuite) {
			},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid body request",
			},
		},

		{
			name:        "Error_Product_NotFound",
			userRole:    "manager",
			path:        "/products/1",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateProduct", mock.Anything, uint(1), mockInput).Return(nil, apperror.ErrNotFound).Once()

			},
			expectedStatusCode: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "product not found",
			},
		},
		{
			name:        "Error_Conflict_SKU_AlreadyExist",
			userRole:    "manager",
			path:        "/products/1",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateProduct", mock.Anything, uint(1), mockInput).Return(nil, apperror.ErrConflict).Once()

			},
			expectedStatusCode: fiber.StatusConflict,
			expectedBody: fiber.Map{

				"code":    "CONFLICT",
				"message": "product sku already",
			},
		},
		{
			name:        "Error_InternalServer",
			userRole:    "manager",
			path:        "/products/1",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateProduct", mock.Anything, uint(1), mockInput).Return(nil, apperror.ErrInternalServer).Once()

			},
			expectedStatusCode: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetUpHandlerTestSuite(t)

			// Create Middleware for mock jwt claims
			ts.App.Use(func(c *fiber.Ctx) error {
				if test.userRole != "" {
					c.Locals("user", &jwtx.Claims{UserID: 1, Username: "Test", Role: test.userRole})
				}
				return c.Next()
			})

			// Create route
			ts.App.Patch("/products/:id", ts.Handler.UpdateProduct)
			test.setup(ts)

			// create body request
			var body []byte
			if value, ok := test.requestBody.(string); ok {
				body = []byte(value)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			// create request
			req := httptest.NewRequest(fiber.MethodPatch, test.path, bytes.NewBuffer(body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// run test
			res, _ := ts.App.Test(req, -1)

			// status code check
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			// tranform data type to byte
			expectedBody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resBody))

		})
	}
}

func TestProductHandler_DeleteProduct(t *testing.T) {
	tests := []struct {
		name               string
		userRole           string
		path               string
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name:     "Success_Delete_Product_With_Manager",
			userRole: "manager",
			path:     "/products/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("DeleteProduct", mock.Anything, uint(1)).Return(nil).Once()
			},
			expectedStatusCode: fiber.StatusNoContent,
		},
		{
			name:               "Error_Forbidden_With_Cashier",
			userRole:           "cashier",
			path:               "/products/1",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusForbidden,
			expectedBody: fiber.Map{

				"code":    "FORBIDDEN",
				"message": "insufficient permission",
			},
		},
		{
			name:               "Error_BadRequest_Invalid_ProductID_Format",
			userRole:           "manager",
			path:               "/products/one",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid productID",
			},
		},
		{
			name:     "Error_Product_NotFound",
			userRole: "manager",
			path:     "/products/99",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("DeleteProduct", mock.Anything, uint(99)).Return(apperror.ErrNotFound).Once()
			},
			expectedStatusCode: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "product not found",
			},
		},
		{
			name:     "Error_InternalServer",
			userRole: "manager",
			path:     "/products/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("DeleteProduct", mock.Anything, uint(1)).Return(apperror.ErrInternalServer).Once()
			},
			expectedStatusCode: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetUpHandlerTestSuite(t)

			// Mock jwt cliams
			ts.App.Use(func(c *fiber.Ctx) error {
				if test.userRole != "" {
					c.Locals("user", &jwtx.Claims{UserID: 1, Username: "Test", Role: test.userRole})
				}
				return c.Next()
			})

			// Create Route
			ts.App.Delete("/products/:id", ts.Handler.DeleteProduct)
			test.setup(ts)

			// Create request
			req := httptest.NewRequest(fiber.MethodDelete, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Run test
			res, _ := ts.App.Test(req, -1)

			// Status code check
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			if test.expectedBody != nil {
				// Tranform body to byte
				expectedBody, _ := json.Marshal(test.expectedBody)
				resBody, _ := io.ReadAll(res.Body)

				assert.JSONEq(t, string(expectedBody), string(resBody))
			}

		})
	}
}

func TestProductHandler_List(t *testing.T) {
	mockProducts := fixtures.ValidListProduct()
	mockQuery := product.ListQuery{
		Search: "",
		Limit:  10,
		Offset: 0,
		Sort:   "ASC",
	}
	mockOutput := createListOutput(mockProducts)
	mockResponse := createListResponse(mockOutput.Items)

	tests := []struct {
		name               string
		path               string
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name: "Success_Get_AllProduct",
			path: "/products/?limit=10&offset=0",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("List", mock.Anything, mockQuery).Return(mockOutput, nil).Once()
			},
			expectedStatusCode: fiber.StatusOK,
			expectedBody:       mockResponse,
		},
		{
			name:               "Error_BadRequest_InvalidParams_Limit",
			path:               "/products?limit=invalid",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid limit request",
			},
		},
		{
			name:               "Error_BadRequest_InvalidParams_Offset",
			path:               "/products?offset=invalid",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid offset request",
			},
		},
		{
			name: "Error_InternalServer",
			path: "/products?limit=10&offset=0&search",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("List", mock.Anything, mockQuery).Return(nil, apperror.ErrInternalServer).Once()
			},
			expectedStatusCode: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetUpHandlerTestSuite(t)

			// Create route
			ts.App.Get("/products", ts.Handler.List)
			test.setup(ts)

			// Create Request
			req := httptest.NewRequest(fiber.MethodGet, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Run test
			res, _ := ts.App.Test(req, -1)

			// Status code check
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			// tranform type to byte
			expectedBody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resBody))
		})
	}
}

func createListOutput(products []*domain.Product) *product.ListOutput {
	items := make([]*product.ItemLite, len(products))
	for i, p := range products {
		items[i] = &product.ItemLite{
			ID:         p.ID,
			Name:       p.Name,
			SKU:        p.SKU,
			Price:      p.Price,
			CategoryID: p.CategoryID,
		}
	}

	return &product.ListOutput{
		Items: items,
		Total: int64(len(items)),
	}
}

func createListResponse(item []*product.ItemLite) *product.ProductListResponse {
	items := make([]*product.LiteProductResponse, len(item))
	for i, item := range item {
		items[i] = &product.LiteProductResponse{
			ID:         item.ID,
			Name:       item.Name,
			SKU:        item.SKU,
			Price:      item.Price,
			CategoryID: item.CategoryID,
		}
	}

	return &product.ProductListResponse{
		Products: items,
		Total:    int64(len(items)),
	}
}
