package auth_test

import (
	"ans-spareparts-api/internal/domain"
	"ans-spareparts-api/internal/features/auth"
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
	Handler     *auth.Handler
	MockService *mocks.AuthService
}

func NewHandlerTestSuite() *HandlerTestSuite {
	return &HandlerTestSuite{}
}

func (ts *HandlerTestSuite) SetupTestSuite(t *testing.T) {
	ts.MockService = mocks.NewAuthService()
	ts.Handler = auth.NewHandler(ts.MockService)
	// สร้าง Fiber app พร้อม Default Error Handler
	ts.App = fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body: "+err.Error())
		},
	})

	t.Cleanup(func() {
		ts.MockService.AssertExpectations(t)
	})
}

func TestAuthHandler_Register(t *testing.T) {
	mockUser := fixtures.ValidUser()
	mockRequest := auth.RegisterRequest{
		Username: mockUser.Username,
		Email:    mockUser.Email,
		Password: mockUser.Password,
	}
	mockInput := auth.RegisterInput{
		Username: mockUser.Username,
		Email:    mockUser.Email,
		Password: mockUser.Password,
	}
	expectedService := &domain.User{
		Username: mockUser.Username,
		Email:    mockUser.Email,
		Password: "",
		Role:     mockUser.Role,
	}

	tests := []struct {
		name           string
		path           string
		requestBody    interface{}
		setup          func(*HandlerTestSuite)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success_Register_Cashier",
			path: "/auth/register",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Register", mock.Anything, mockInput).Return(expectedService, nil).Once()
			},
			requestBody:    mockRequest,
			expectedStatus: fiber.StatusOK,
			expectedBody: fiber.Map{
				"message": "register successfull",
			},
		},
		{
			name: "Error_BadRequest_Invalid_RequestBody",
			path: "/auth/register",
			setup: func(hts *HandlerTestSuite) {
			},
			requestBody:    `{"username":"missing double-qoute}`,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{
				"code":    "BAD_REQUEST",
				"message": "invalid json body",
			},
		},
		{
			name: "Error_Conflict_User_AlreadyExist",
			path: "/auth/register",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Register", mock.Anything, mockInput).Return(nil, apperror.ErrConflict).Once()
			},
			requestBody:    mockRequest,
			expectedStatus: fiber.StatusConflict,
			expectedBody: fiber.Map{
				"code":    "CONFLICT",
				"message": "user already exists",
			},
		},
		{
			name: "Error_BadRequest_InvalidInput",
			path: "/auth/register",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Register", mock.Anything, mockInput).Return(nil, apperror.ErrInvalidInput).Once()
			},
			requestBody:    mockRequest,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{
				"code":    "BAD_REQUEST",
				"message": "invalid input",
			},
		},
		{
			name: "Error_InternalServer",
			path: "/auth/register",
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Register", mock.Anything, mockInput).Return(nil, apperror.ErrInternalServer).Once()
			},
			requestBody:    mockRequest,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error occurred",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetupTestSuite(t)

			// สร้าง Route
			ts.App.Post(test.path, ts.Handler.Register)
			test.setup(ts)

			var body []byte
			if value, ok := test.requestBody.(string); ok {
				body = []byte(value)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			// สร้าง request
			req := httptest.NewRequest(fiber.MethodPost, test.path, bytes.NewBuffer(body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Run test
			res, _ := ts.App.Test(req, -1)

			// Status check
			assert.Equal(t, test.expectedStatus, res.StatusCode)

			// เปลี่ยน res และ expectedBpdy เป็น byte
			resBody, _ := io.ReadAll(res.Body)
			expectedBody, _ := json.Marshal(test.expectedBody)

			// เปลียบเทียบผลลัพท์
			assert.JSONEq(t, string(expectedBody), string(resBody))
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	mockUser := fixtures.ValidUser()
	mockInput := auth.LoginInput{
		Username: mockUser.Username,
		Password: mockUser.Password,
	}
	mockRequest := auth.LoginRequest{
		Username: mockUser.Username,
		Password: mockUser.Password,
	}
	expectedBody := fiber.Map{
		"token": "token",
	}

	tests := []struct {
		name           string
		path           string
		requestBody    interface{}
		setup          func(*HandlerTestSuite)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "Success_Login",
			path:        "/auth/login",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Login", mock.Anything, mockInput).Return(mockUser, "token", nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   expectedBody,
		},
		{
			name:           "Error_BadRequest_Invalid_RequestBody",
			path:           "/auth/login",
			requestBody:    `{"username":"missing double quote}`,
			setup:          func(hts *HandlerTestSuite) {},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody: fiber.Map{
				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		},
		{
			name:        "Error_Unautherized_WrongPassword",
			path:        "/auth/login",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Login", mock.Anything, mockInput).Return(nil, "", apperror.ErrUnauthorized).Once()

			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody: fiber.Map{

				"code":    "UNAUTHERIZED",
				"message": "invalid credentials",
			},
		},
		{
			name:        "Error_InternalServer",
			path:        "/auth/login",
			requestBody: mockRequest,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Login", mock.Anything, mockInput).Return(nil, "", apperror.ErrInternalServer).Once()
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
			ts.SetupTestSuite(t)

			// setup route
			ts.App.Post(test.path, ts.Handler.Login)
			test.setup(ts)

			// create body
			var body []byte
			if value, ok := test.requestBody.(string); ok {
				body = []byte(value)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			// create request
			req := httptest.NewRequest(fiber.MethodPost, test.path, bytes.NewBuffer(body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// test start
			res, _ := ts.App.Test(req, -1)

			// code status check
			assert.Equal(t, test.expectedStatus, res.StatusCode)

			// map response body and expected body to byte
			resBody, _ := io.ReadAll(res.Body)
			expectedBody, _ := json.Marshal(test.expectedBody)

			assert.JSONEq(t, string(expectedBody), string(resBody))
		})
	}
}

func TestAuthenHandler_Logout(t *testing.T) {
	mockToken := "valid-token-123"
	headerValue := "Bearer " + mockToken // ค่าที่ส่งไปที่ Header
	mockPath := "/auth/logout"

	tests := []struct {
		name           string
		path           string
		headerValue    string
		setup          func(*HandlerTestSuite)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "Success_Logout",
			path:        mockPath,
			headerValue: headerValue,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Logout", mock.Anything, mockToken).Return(nil).Once()
			},
			expectedStatus: fiber.StatusOK,
			expectedBody: fiber.Map{
				"message": "logout success",
			},
		},
		{
			name:           "Error_Unautherized_Missing_Token",
			path:           mockPath,
			headerValue:    "",
			setup:          func(hts *HandlerTestSuite) {},
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody: fiber.Map{

				"code":    "UNAUTHERIZED",
				"message": "missing token",
			},
		},
		{
			name:        "Error_InternalServer",
			path:        mockPath,
			headerValue: headerValue,
			setup: func(hts *HandlerTestSuite) {
				hts.MockService.On("Logout", mock.Anything, mockToken).Return(apperror.ErrInternalServer).Once()
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "failed to logout",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewHandlerTestSuite()
			ts.SetupTestSuite(t)

			// Create Route
			ts.App.Post(test.path, ts.Handler.Logout)
			test.setup(ts)

			// Create Request
			req := httptest.NewRequest(fiber.MethodPost, test.path, nil)
			req.Header.Set("Authorization", test.headerValue)

			// Run test
			res, _ := ts.App.Test(req, -1)

			assert.Equal(t, test.expectedStatus, res.StatusCode)

			resBody, _ := io.ReadAll(res.Body)
			expectedBody, _ := json.Marshal(test.expectedBody)

			assert.JSONEq(t, string(expectedBody), string(resBody))

		})
	}

}
