package inventory_test

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/inventory"
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
	MockService *mocks.InventoryService
	Handler     *inventory.Handler
	App         *fiber.App
}

func NewHandlerTestSuite() *HandlerTestSuite {
	return &HandlerTestSuite{}
}

func (ts *HandlerTestSuite) SetUpHandlerTestSuite(t *testing.T) {
	ts.MockService = mocks.NewInventoryService()
	ts.Handler = inventory.NewHandler(ts.MockService)

	ts.App = fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body: "+err.Error())
		},
	},
	)

	t.Cleanup(func() {
		ts.MockService.AssertExpectations(t)
	})

}

func TestInventoryHandler_GetInventoryByID(t *testing.T) {
	mockInv := fixtures.ValidInventory()
	mockItem := &inventory.Item{
		ID:        mockInv.ID,
		ProductID: mockInv.ProductID,
		Quantity:  mockInv.Quantity,
	}
	mockRes := &inventory.InventoryResponse{
		ID:        mockInv.ID,
		ProductID: mockInv.ProductID,
		Quantity:  mockInv.Quantity,
	}

	tests := []struct {
		name           string
		path           string
		setup          func(*HandlerTestSuite)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success_GetInventory_By_ID",
			path: "/inventories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetInventoryByID", mock.Anything, uint(1)).Return(mockItem, nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   mockRes,
		},
		{
			name: "Error_BadRequest_Invalid_InventoryID",
			path: "/inventories/one",
			setup: func(hts *HandlerTestSuite) {
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid inventory id request",
			},
		},
		{
			name: "Error_Inventory_NotFound",
			path: "/inventories/99",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetInventoryByID", mock.Anything, uint(99)).Return(nil, apperror.ErrNotFound).Once()
			},
			expectedStatus: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "inventory not found",
			},
		},
		{
			name: "Error_InternalServer",
			path: "/inventories/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetInventoryByID", mock.Anything, uint(1)).Return(nil, apperror.ErrInternalServer).Once()

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
			// setup test suite
			ts := NewHandlerTestSuite()
			ts.SetUpHandlerTestSuite(t)

			// Create route path
			ts.App.Get("/inventories/:id", ts.Handler.GetInventoryByID)
			test.setup(ts)

			// Create Request
			req := httptest.NewRequest(fiber.MethodGet, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Start test
			res, _ := ts.App.Test(req, -1)

			// Status code check
			assert.Equal(t, test.expectedStatus, res.StatusCode)

			// แปลง expected body และ res body เป็น byte
			expectedBody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resBody))
		})
	}
}

func TestInventoryHandler_GetInventoryByProductID(t *testing.T) {
	mockInv := fixtures.ValidInventory()
	mockItem := &inventory.Item{
		ID:        mockInv.ID,
		ProductID: mockInv.ProductID,
		Quantity:  mockInv.Quantity,
	}
	mockRes := &inventory.InventoryResponse{
		ID:        mockInv.ID,
		ProductID: mockInv.ProductID,
		Quantity:  mockInv.Quantity,
	}

	tests := []struct {
		name           string
		path           string
		setup          func(*HandlerTestSuite)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success_GetInventory_By_ProductID",
			path: "/product/1/inventory",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetInventoryByID", mock.Anything, uint(1)).Return(mockItem, nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   mockRes,
		},
		{
			name: "Error_BadRequest_Invalid_ProductID",
			path: "/product/one/inventory",
			setup: func(hts *HandlerTestSuite) {
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid inventory id request",
			},
		},
		{
			name: "Error_Inventory_NotFound",
			path: "/product/99/inventory",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetInventoryByID", mock.Anything, uint(99)).Return(nil, apperror.ErrNotFound).Once()
			},
			expectedStatus: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "inventory not found",
			},
		},
		{
			name: "Error_InternalServer",
			path: "/product/1/inventory",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("GetInventoryByID", mock.Anything, uint(1)).Return(nil, apperror.ErrInternalServer).Once()

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
			// setup test suite
			ts := NewHandlerTestSuite()
			ts.SetUpHandlerTestSuite(t)

			// Create route path
			ts.App.Get("/product/:id/inventory", ts.Handler.GetInventoryByID)
			test.setup(ts)

			// Create Request
			req := httptest.NewRequest(fiber.MethodGet, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Start test
			res, _ := ts.App.Test(req, -1)

			// Status code check
			assert.Equal(t, test.expectedStatus, res.StatusCode)

			// แปลง expected body และ res body เป็น byte
			expectedBody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resBody))
		})
	}

}

func TestInventoryHandler_UpdateQuantity(t *testing.T) {
	mockInv := fixtures.ValidInventory()
	mockInput := inventory.UpdateQuantityInput{
		ProductID: mockInv.ID,
		Quantity:  1,
	}
	mockItem := &inventory.Item{
		ID:        mockInv.ID,
		ProductID: mockInv.ProductID,
		Quantity:  2,
	}
	mockRequest := &inventory.UpdateQuantityRequest{
		ProductID: mockInv.ID,
		Quantity:  1,
	}
	mockResponse := &inventory.InventoryResponse{
		ID:        mockInv.ID,
		ProductID: mockInput.ProductID,
		Quantity:  2,
	}

	tests := []struct {
		name               string
		path               string
		userRole           string
		requestBody        interface{}
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name:        "Success_Incread_Quantity_+1",
			userRole:    "manager",
			path:        "/inventories/1",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateQuantity", mock.Anything, uint(1), mockInput).Return(mockItem, nil).Once()
			},
			expectedStatusCode: fiber.StatusOK,
			expectedBody:       mockResponse,
		},
		{
			name:     "Success_Reduce_Quantity_-1",
			userRole: "manager",
			path:     "/inventories/1",
			requestBody: inventory.UpdateQuantityInput{
				ProductID: 1,
				Quantity:  -1,
			},
			setup: func(hts *HandlerTestSuite) {
				input := inventory.UpdateQuantityInput{
					ProductID: 1,
					Quantity:  -1,
				}
				item := &inventory.Item{
					ID:        1,
					ProductID: 1,
					Quantity:  0,
				}
				hts.MockService.On("UpdateQuantity", mock.Anything, uint(1), input).Return(item, nil).Once()
			},
			expectedStatusCode: fiber.StatusOK,
			expectedBody: &inventory.InventoryResponse{
				ID:        mockInv.ID,
				ProductID: mockInput.ProductID,
				Quantity:  0,
			},
		},
		{
			name:     "Error_Unprocessable_Insufficient_Stock",
			userRole: "manager",
			path:     "/inventories/1",
			requestBody: inventory.UpdateQuantityInput{
				ProductID: 1,
				Quantity:  -2,
			},
			setup: func(hts *HandlerTestSuite) {
				input := inventory.UpdateQuantityInput{
					ProductID: 1,
					Quantity:  -2,
				}

				hts.MockService.On("UpdateQuantity", mock.Anything, uint(1), input).Return(nil, apperror.ErrInsufficientStock).Once()
			},
			expectedStatusCode: fiber.StatusUnprocessableEntity,
			expectedBody: fiber.Map{

				"code":    "UNPROCESSABLE",
				"message": "Insufficient Stock",
			},
		},
		{
			name:        "Error_Forbidden_UpdateQuantity_With_Cashier",
			userRole:    "cashier",
			path:        "/inventories/1",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
			},
			expectedStatusCode: fiber.StatusForbidden,
			expectedBody: fiber.Map{

				"code":    "FORBIDDEN",
				"message": "Insufficeint permission",
			},
		},
		{
			name:               "Error_BadRequest_Invalid_InventoryID",
			userRole:           "manager",
			path:               "/inventories/one",
			requestBody:        mockRequest,
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid inventoryid",
			},
		},
		{
			name:               "Error_BadRequest_Invalid_BodyRequest",
			userRole:           "manager",
			path:               "/inventories/1",
			requestBody:        `{"name": "missing single quote}`,
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid body request",
			},
		},
		{
			name:        "Error_Inventory_Notfound",
			userRole:    "manager",
			path:        "/inventories/99",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateQuantity", mock.Anything, uint(99), mockInput).Return(nil, apperror.ErrNotFound).Once()
			},
			expectedStatusCode: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "inventory for product not found",
			},
		},

		{
			name:        "Error_InternalServer",
			userRole:    "manager",
			path:        "/inventories/1",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("UpdateQuantity", mock.Anything, uint(1), mockInput).Return(nil, apperror.ErrInternalServer).Once()
			},
			expectedStatusCode: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occuted",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetUpHandlerTestSuite(t)

			// Create mock user Context
			ts.App.Use(func(c *fiber.Ctx) error {
				if test.userRole != "" {
					c.Locals("user", &jwtx.Claims{Username: "Test", Role: test.userRole, UserID: 1})
				}
				return c.Next()
			})

			// Create Route
			ts.App.Patch("/inventories/:id", ts.Handler.UpdateQuantity)
			test.setup(ts)

			// Create Body []byte
			var body []byte
			if value, ok := test.requestBody.(string); ok {
				body = []byte(value)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			// Create Request
			req := httptest.NewRequest(fiber.MethodPatch, test.path, bytes.NewBuffer(body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Run test
			res, _ := ts.App.Test(req, -1)

			// Status check
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			// Change expected body and response body to []byte
			expectedBody, _ := json.Marshal(test.expectedBody)
			resonseBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resonseBody))
		})
	}
}

func TestInventoryHandler_List(t *testing.T) {
	mockList := fixtures.ValidListInventory()
	mockQuery := inventory.ListQuery{
		Limit:  10,
		Offset: 0,
		Sort:   "ASC",
	}
	mockOutput := createListOutput(mockList)
	mockResponse := createListResponse(mockOutput.Items)

	tests := []struct {
		name               string
		userRole           string
		path               string
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name:     "Success_Get_All_Inventory",
			userRole: "manager",
			path:     "/inventories?limit=10&offset=0&sort=ASC",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("List", mock.Anything, mockQuery).Return(mockOutput, nil).Once()
			},
			expectedStatusCode: fiber.StatusOK,
			expectedBody:       mockResponse,
		},
		{
			name:               "Error_Forbidden_By_Cashier",
			userRole:           "cashier",
			path:               "/inventories?limit=10&offset=0&sort=ASC",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusForbidden,
			expectedBody: fiber.Map{

				"code":    "FORBIDDEN",
				"message": "insufficient permission",
			},
		},
		{
			name:               "Error_BadRequest_Invalid_Limit",
			userRole:           "manager",
			path:               "/inventories?limit=zero0",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid limit request",
			},
		},

		{
			name:               "Error_BadRequest_InvalidParams_Offset",
			userRole:           "manager",
			path:               "/inventories?offset=invalid",
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedBody: fiber.Map{

				"code":    "BAD_REQUEST",
				"message": "invalid offset request",
			},
		},
		{
			name:     "Error_InternalServer",
			userRole: "manager",
			path:     "/inventories?limit=10&offset=0&search=test",
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

			ts.App.Use(func(c *fiber.Ctx) error {
				if test.userRole != "" {
					c.Locals("user", &jwtx.Claims{UserID: 1, Username: "Test", Role: test.userRole})
				}
				return c.Next()
			})

			// Create Route
			ts.App.Get("/inventories", ts.Handler.List)
			test.setup(ts)

			// Create request
			req := httptest.NewRequest(fiber.MethodGet, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// run test
			res, _ := ts.App.Test(req, -1)

			// status code check
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			// change type body to byte
			expectedbody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedbody), string(resBody))
		})
	}

}

// -- helper
func createListOutput(inventories []*domain.Inventory) *inventory.ListOutput {
	items := make([]*inventory.Item, len(inventories))
	for i, inv := range inventories {
		items[i] = &inventory.Item{
			ID:        inv.ID,
			ProductID: inv.ProductID,
			Quantity:  inv.Quantity,
		}
	}
	return &inventory.ListOutput{
		Items: items,
		Total: int64(len(items)),
	}
}

func createListResponse(items []*inventory.Item) *inventory.InventoryListResponse {
	list := make([]*inventory.InventoryResponse, len(items))
	for i, item := range items {
		list[i] = &inventory.InventoryResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return &inventory.InventoryListResponse{
		Inventories: list,
		Total:       int64(len(list)),
	}
}
