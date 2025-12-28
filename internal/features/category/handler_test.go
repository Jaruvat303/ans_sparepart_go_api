package category_test

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/category"
	"ans-spareparts-api/internal/infra/jwtx"
	mocks "ans-spareparts-api/internal/mock"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/response"
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
	Handler     *category.Handler
	MockService *mocks.CategoryService
}

func NewHandlerTestSuite() *HandlerTestSuite {
	return &HandlerTestSuite{}
}

func (ts *HandlerTestSuite) SetupTest(t *testing.T) {

	ts.MockService = mocks.NewMockCategoryService()

	// สร้าง Fiber app พร้อม Default Error Handler
	ts.App = fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body: "+err.Error())
		},
	})

	ts.Handler = category.NewHandler(ts.MockService)

	// cleanup
	t.Cleanup(func() {
		ts.MockService.AssertExpectations(t)
	})
}

func TestCategoryHandler_CreateCategory(t *testing.T) {
	validCategory := fixtures.ValidCategory()
	expectedCategory := &category.Item{
		ID:   validCategory.ID,
		Name: validCategory.Name,
	}
	mockrequest := category.CategoryRequest{
		Name: validCategory.Name,
	}

	tests := []struct {
		name           string
		userRole       string
		path           string
		requestBody    interface{}
		setup          func(*HandlerTestSuite)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "Success_Created_By_Admin",
			path:        "/categories",
			userRole:    "admin",
			requestBody: mockrequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("CreateCategory", mock.Anything, mockrequest).Return(expectedCategory, nil).Once()
			},
			expectedStatus: fiber.StatusCreated,
			expectedBody: fiber.Map{
				"ID":   1,
				"Name": "Wheel",
			},
		},
		{
			name:           "Forbidden_By_Cashier_User",
			path:           "/categories",
			userRole:       "cashier",
			requestBody:    mockrequest,
			setup:          func(hts *HandlerTestSuite) {}, // ไม่ทำงาน user role ไม่ผ่าน
			expectedStatus: fiber.StatusForbidden,
			expectedBody: fiber.Map{
				"code":    "FORBIDDEN",
				"message": "Insufficient permissions",
			},
		},
		{
			name:        "Confliect_Duplicate_Name",
			userRole:    "admin",
			path:        "/categories",
			requestBody: mockrequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("CreateCategory", mock.Anything, mockrequest).Return(nil, apperror.ErrConflict).Once()
			},
			expectedStatus: fiber.StatusConflict,
			expectedBody: fiber.Map{

				"code":    "CONFLICT",
				"message": "category name already exist",
			},
		},
		{
			name:           "BadRequest_Invalid_JSON",
			path:           "/categories",
			userRole:       "manager",
			requestBody:    `{"name": "missing-doublequote}`,
			setup:          func(hts *HandlerTestSuite) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		},
		{
			name:        "Internal_dbError",
			path:        "/categories",
			userRole:    "manager",
			requestBody: mockrequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("CreateCategory", mock.Anything, mockrequest).Return(nil, apperror.ErrInternalServer).Once()
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup test suite
			ts := NewHandlerTestSuite()
			ts.SetupTest(t)

			// Middle ware จำลอง user context / jwt cliams
			ts.App.Use(func(c *fiber.Ctx) error {
				if test.userRole != "" {
					c.Locals("user", &jwtx.Claims{Username: "Test", Role: test.userRole, UserID: 1})
				}
				return c.Next()
			})

			//
			ts.App.Post(test.path, ts.Handler.CreateCategory)
			test.setup(ts)

			// เตรียม payload
			var body []byte
			if b, ok := test.requestBody.(string); ok {
				body = []byte(b)
			} else {
				// แปลง body ที่ไม่ใช่ string เป็น []byte
				body, _ = json.Marshal(test.requestBody)
			}

			// สร้าง Request
			req := httptest.NewRequest(fiber.MethodPost, test.path, bytes.NewBuffer(body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Run test
			res, _ := ts.App.Test(req, -1)

			// Status code cheack
			assert.Equal(t, test.expectedStatus, res.StatusCode)

			// อ่านข้อมูลที่ Handler ตอบกลับมา (เป็น Byte )
			respBody, _ := io.ReadAll(res.Body)
			// แปลง expectedbody ให้เป็น Byte
			expectedJson, _ := json.Marshal(test.expectedBody)
			// ใช้ JSONEq เพื่อตรวจสอบโตรงสร้าง Json ตรงกันหรือไม่
			assert.JSONEq(t, string(expectedJson), string(respBody))

		})
	}
}

func TestCategoryHandler_GetCategory(t *testing.T) {
	validCategory := fixtures.ValidCategory()
	expectedCategory := &category.Item{
		ID:   validCategory.ID,
		Name: validCategory.Name,
	}
	mockResponse := fiber.Map{
		"ID":   1,
		"Name": "Wheel",
	}

	tests := []struct {
		name           string
		path           string
		setup          func(*HandlerTestSuite)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success_Get_Category_By_ID",
			path: "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetCategoryByID", mock.Anything, uint(1)).Return(expectedCategory, nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   mockResponse,
		},
		{
			name: "Error_BadRequest_InvalidIDFormat",
			path: "/categories/invalidID",
			setup: func(hts *HandlerTestSuite) {

			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid category id",
			},
		},
		{
			name: "Error_Category_NotFound",
			path: "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetCategoryByID", mock.Anything, uint(1)).Return(nil, apperror.ErrNotFound).Once()
			},
			expectedStatus: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "category not found",
			},
		},
		{
			name: "Error_InternalServerError",
			path: "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetCategoryByID", mock.Anything, uint(1)).Return(nil, apperror.ErrInternalServer).Once()
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetupTest(t)

			ts.App.Get("categories/:id", ts.Handler.GetCategory)
			test.setup(ts)

			req := httptest.NewRequest(fiber.MethodGet, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			res, _ := ts.App.Test(req, -1)

			// status code check
			assert.Equal(t, test.expectedStatus, res.StatusCode)

			reqBody, _ := io.ReadAll(res.Body)
			expectedJson, _ := json.Marshal(test.expectedBody)
			assert.JSONEq(t, string(expectedJson), string(reqBody))

		})
	}

}

func TestCategoryHandler_UpdateCategory(t *testing.T) {
	validCategory := fixtures.ValidCategory()
	expectedService := &category.Item{
		ID:   validCategory.ID,
		Name: validCategory.Name,
	}
	mockRequest := category.CategoryRequest{
		Name: "Wheel",
	}
	expectedBody := fiber.Map{
		"ID":   1,
		"Name": "Wheel",
	}

	tests := []struct {
		name           string
		userRole       string
		path           string
		setup          func(*HandlerTestSuite)
		requestBody    interface{}
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:     "Success_Update_Category",
			userRole: "manager",
			path:     "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateCategory", mock.Anything, uint(1), mockRequest).Return(expectedService, nil).Once()
			},
			requestBody:    mockRequest,
			expectedStatus: fiber.StatusOK,
			expectedBody:   expectedBody,
		},
		{
			name:           "Error_Forbidden_By_Cashier",
			userRole:       "cashier",
			path:           "/categories/1",
			setup:          func(hts *HandlerTestSuite) {},
			requestBody:    mockRequest,
			expectedStatus: fiber.StatusForbidden,
			expectedBody: fiber.Map{

				"code":    "FORBIDDEN",
				"message": "insufficient permissions",
			},
		},
		{
			name:           "Error_BadRequest_InvalidCateogoryID",
			userRole:       "manager",
			path:           "/categories/one",
			requestBody:    mockRequest,
			setup:          func(hts *HandlerTestSuite) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid categoryID",
			},
		},
		{
			name:           "Error_BadRequest_Invalid_RequestBody",
			userRole:       "manager",
			path:           "/categories/1",
			setup:          func(hts *HandlerTestSuite) {},
			requestBody:    `{"name": "missing-doublequote}`,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invaid request body",
			},
		},
		{
			name:     "Error_Category_NotFound",
			userRole: "manager",
			path:     "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateCategory", mock.Anything, uint(1), mockRequest).Return(nil, apperror.ErrNotFound).Once()
			},
			requestBody:    mockRequest,
			expectedStatus: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "category not found",
			},
		},
		{
			name:     "Error_InternalServer",
			userRole: "manager",
			path:     "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateCategory", mock.Anything, uint(1), mockRequest).Return(nil, apperror.ErrInternalServer).Once()
			},
			requestBody:    mockRequest,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetupTest(t)

			ts.App.Use(func(ctx *fiber.Ctx) error {
				if test.userRole != "" {
					ctx.Locals("user", &jwtx.Claims{UserID: 1, Username: "Test", Role: test.userRole})
				}
				return ctx.Next()
			})

			// สร้าง route
			ts.App.Put("/categories/:id", ts.Handler.UpdateCategory)
			test.setup(ts)

			// เตรียม payload
			var body []byte
			if value, ok := test.requestBody.(string); ok {
				body = []byte(value)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			// สร้าง Request
			req := httptest.NewRequest(fiber.MethodPut, test.path, bytes.NewBuffer(body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Run test
			res, _ := ts.App.Test(req, -1)

			// Status code check
			assert.Equal(t, test.expectedStatus, res.StatusCode)

			// รับ response body เป็น Byte
			respBody, _ := io.ReadAll(res.Body)
			expectedJson, _ := json.Marshal(test.expectedBody)
			assert.JSONEq(t, string(expectedJson), string(respBody))

		})
	}
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		path           string
		setup          func(*HandlerTestSuite)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:     "Success_Delete_Category_With_Admin",
			userRole: "admin",
			path:     "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("DeleteCategory", mock.Anything, uint(1)).Return(nil).Once()
			},
			expectedStatus: fiber.StatusNoContent,
		},
		{
			name:     "Error_Forbbiden_With_Cashier",
			userRole: "cashier",
			path:     "/categories/1",
			setup: func(hts *HandlerTestSuite) {
			},
			expectedStatus: fiber.StatusForbidden,
			expectedBody: fiber.Map{

				"code":    "FORBIDDEN",
				"message": "Insufficient permissions",
			},
		},
		{
			name:     "Error_BadRequest_Invalid_CategoryID",
			userRole: "admin",
			path:     "/categories/one",
			setup: func(hts *HandlerTestSuite) {

			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid category id",
			},
		},
		{
			name:     "Error_Category_NotFound",
			userRole: "admin",
			path:     "/categories/99",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("DeleteCategory", mock.Anything, uint(99)).Return(apperror.ErrNotFound).Once()
			},
			expectedStatus: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "category not found",
			},
		},
		{
			name:     "Error_Conflict_Category_In_Use",
			userRole: "admin",
			path:     "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("DeleteCategory", mock.Anything, uint(1)).Return(apperror.ErrConflict).Once()
			},
			expectedStatus: fiber.StatusConflict,
			expectedBody: fiber.Map{

				"code":    "CONFLICT",
				"message": "category is in use",
			},
		},
		{
			name:     "Error_InternalServer",
			userRole: "admin",
			path:     "/categories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("DeleteCategory", mock.Anything, uint(1)).Return(apperror.ErrInternalServer).Once()
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetupTest(t)

			ts.App.Use(func(ctx *fiber.Ctx) error {
				if test.userRole != "" {
					ctx.Locals("user", &jwtx.Claims{UserID: 1, Username: "Test", Role: test.userRole})
				}
				return ctx.Next()
			})

			// สร้าง route
			ts.App.Delete("categories/:id", ts.Handler.DeleteCategory)
			test.setup(ts)

			// สร้าง Request
			req := httptest.NewRequest(fiber.MethodDelete, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			res, _ := ts.App.Test(req, -1)

			assert.Equal(t, test.expectedStatus, res.StatusCode)

			if test.expectedBody != nil {
				// แปลง response เป็น byte
				resBody, _ := io.ReadAll(res.Body)
				// แปลง expected Body เป็น byte
				expectedBody, _ := json.Marshal(test.expectedBody)
				// เปรียบเทียบ expected กับ response
				assert.JSONEq(t, string(expectedBody), string(resBody))
			}
		})
	}
}

func TestCategoryHandler_List(t *testing.T) {
	mockList := fixtures.ValidListCategory()
	mockQuery := category.ListQuery{
		Limit:  5,
		Offset: 0,
		Search: "test",
	}
	mockListOutput := createListOutput(mockList)
	mockListResponse := createListResponse(mockListOutput.Items)

	tests := []struct {
		name               string
		path               string
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name: "Success_Get_All_Category",
			path: "/categories?limit=5&offset=0&search=test",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("List", mock.Anything, mockQuery).Return(mockListOutput, nil).Once()
			},
			expectedStatusCode: fiber.StatusOK,
			expectedBody:       mockListResponse,
		},
		{
			name:               "Error_BadRequest_InvalidParams_Limit",
			path:               "/categories?limit=invalid",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid limit request",
			},
		},
		{
			name:               "Error_BadRequest_InvalidParams_Offset",
			path:               "/categories?offset=invalid",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid offset request",
			},
		},
		{
			name: "Error_InternalServer",
			path: "/categories?limit=5&offset=0&search=test",
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
			ts.SetupTest(t)

			// Create Route
			ts.App.Get("/categories", ts.Handler.List)
			test.setup(ts)

			// Create request
			req := httptest.NewRequest(fiber.MethodGet, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// run test
			res, _ := ts.App.Test(req, -1)

			// status check
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			// change data type
			expectedBody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resBody))
		})
	}
}

// --- Halper function

func createListOutput(categories []*domain.Category) *category.ListOutput {
	list := make([]*category.Item, len(categories))
	for i, cat := range categories {
		list[i] = &category.Item{
			ID:   cat.ID,
			Name: cat.Name,
		}
	}
	return &category.ListOutput{
		Items: list,
		Total: int64(len(list)),
	}
}

func createListResponse(items []*category.Item) *category.CategoryListResponse {
	categories := make([]*category.CategoryResponse, len(items))
	for i, item := range items {
		categories[i] = &category.CategoryResponse{
			ID:   item.ID,
			Name: item.Name,
		}
	}

	return &category.CategoryListResponse{
		Categories: categories,
		Total:      int64(len(categories)),
	}
}
