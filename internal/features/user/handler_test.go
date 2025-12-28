package user_test

import (
	"ans-spareparts-api/internal/features/user"
	"ans-spareparts-api/internal/infra/jwtx"
	mocks "ans-spareparts-api/internal/mock"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/response"
	"ans-spareparts-api/pkg/testutil/fixtures"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type HandlerTestSuite struct {
	MockUserService *mocks.UserService
	MockAuthService *mocks.AuthService
	Handler         *user.Handler
	App             *fiber.App
}

func NewHandlerTestSuite() *HandlerTestSuite {
	return &HandlerTestSuite{}
}

func (ts *HandlerTestSuite) SetupHandlerTestSuite(t *testing.T) {
	ts.MockUserService = mocks.NewUserService()
	ts.MockAuthService = mocks.NewAuthService()
	ts.Handler = user.NewHandler(ts.MockUserService, ts.MockAuthService)

	ts.App = fiber.New(
		fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body: "+err.Error())
			},
		},
	)
}

func TestUserHandler_GetUserProfile(t *testing.T) {
	mockUser := fixtures.ValidUser()
	mockItem := &user.Item{
		ID:       mockUser.ID,
		Username: mockUser.Username,
		Email:    mockUser.Email,
		Role:     mockUser.Role,
		IsActive: mockUser.IsActive,
	}
	mockResponse := &user.UserResponse{
		ID:       mockUser.ID,
		Username: mockUser.Username,
		Email:    mockUser.Email,
		Role:     mockUser.Role,
		IsActive: mockUser.IsActive,
	}

	tests := []struct {
		name               string
		userID             uint
		path               string
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name:   "Success_Get_UserProfile",
			userID: 1,
			path:   "/user/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockUserService.On("GetUserProfile", mock.Anything, uint(1)).Return(mockItem, nil).Once()
			},
			expectedStatusCode: fiber.StatusOK,
			expectedBody:       mockResponse,
		},
		{
			name:   "Error_User_NotFound",
			userID: 99,
			path:   "/user/99",
			setup: func(hts *HandlerTestSuite) {
				hts.MockUserService.On("GetUserProfile", mock.Anything, uint(99)).Return(nil, apperror.ErrNotFound).Once()
			},
			expectedStatusCode: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "user not found",
			},
		},
		{
			name:   "Error_InternalServer",
			userID: 1,
			path:   "/user/1",
			setup: func(hts *HandlerTestSuite) {
				hts.MockUserService.On("GetUserProfile", mock.Anything, uint(1)).Return(nil, apperror.ErrInternalServer).Once()
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
			ts.SetupHandlerTestSuite(t)

			// Middleware for mock jwt claims
			ts.App.Use(func(c *fiber.Ctx) error {
				c.Locals("user", &jwtx.Claims{UserID: test.userID, Username: "Test", Role: "cashier"})
				return c.Next()
			})

			// create route
			ts.App.Get("/user/:id", ts.Handler.GetProfile)
			test.setup(ts)

			// create request
			req := httptest.NewRequest(fiber.MethodGet, test.path, nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// run test
			res, _ := ts.App.Test(req, -1)

			// Status Check
			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			// change body type
			expectedBody, _ := json.Marshal(test.expectedBody)
			resBody, _ := io.ReadAll(res.Body)

			assert.JSONEq(t, string(expectedBody), string(resBody))
		})
	}
}

func TestUserHandler_DeleteProfile(t *testing.T) {
	mockToken := "valid-token-123"
	headerValue := "Bearer " + mockToken // ค่าที่ส่งไปที่ Header
	tests := []struct {
		name               string
		path               string
		haederValue        string
		userID             uint
		setup              func(*HandlerTestSuite)
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name:        "Success_Delete_Profile",
			path:        "/user/1",
			haederValue: headerValue,
			userID:      1,
			setup: func(hts *HandlerTestSuite) {
				hts.MockUserService.On("DeleteUser", mock.Anything, uint(1)).Return(nil).Once()
				hts.MockAuthService.On("Logout", mock.Anything, mockToken).Return(nil).Once()
			},
			expectedStatusCode: fiber.StatusNoContent,
		},
		{
			name:               "Error_Unauthorized_Missing_Token",
			path:               "/user/1",
			haederValue:        "",
			userID:             1,
			setup:              func(hts *HandlerTestSuite) {},
			expectedStatusCode: fiber.StatusUnauthorized,
			expectedBody: fiber.Map{

				"code":    "UNAUTHORIZED",
				"message": "missing token",
			},
		},
		{
			name:        "Error_Delete_User_NotFound",
			path:        "/user/99",
			userID:      99,
			haederValue: headerValue,
			setup: func(hts *HandlerTestSuite) {
				hts.MockUserService.On("DeleteUser", mock.Anything, uint(99)).Return(apperror.ErrNotFound).Once()

			},
			expectedStatusCode: fiber.StatusNotFound,
			expectedBody: fiber.Map{

				"code":    "NOT_FOUND",
				"message": "user not found",
			},
		},
		{
			name:        "Error_Delete_User_InternalServer",
			path:        "/user/1",
			haederValue: headerValue,
			userID:      1,
			setup: func(hts *HandlerTestSuite) {
				hts.MockUserService.On("DeleteUser", mock.Anything, uint(1)).Return(apperror.ErrInternalServer).Once()
			},
			expectedStatusCode: fiber.StatusInternalServerError,
			expectedBody: fiber.Map{

				"code":    "INTERNAL_ERROR",
				"message": "internal server occured",
			},
		},
		{
			name:        "Error_Logout_Invalid_Token",
			path:        "/user/1",
			haederValue: headerValue,
			userID:      1,
			setup: func(hts *HandlerTestSuite) {
				hts.MockUserService.On("DeleteUser", mock.Anything, uint(1)).Return(nil).Once()
				hts.MockAuthService.On("Logout", mock.Anything, mockToken).Return(apperror.ErrInvalidToken).Once()

			},
			expectedStatusCode: fiber.StatusUnauthorized,
			expectedBody: fiber.Map{

				"code":    "UNAUTHORIZED",
				"message": "invalid token",
			},
		},
		{
			name:        "Error_Logout_InternalServer",
			path:        "/user/1",
			haederValue: headerValue,
			userID:      1,
			setup: func(hts *HandlerTestSuite) {
				hts.MockUserService.On("DeleteUser", mock.Anything, uint(1)).Return(nil).Once()
				hts.MockAuthService.On("Logout", mock.Anything, mockToken).Return(apperror.ErrInternalServer).Once()

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
			ts.SetupHandlerTestSuite(t)

			// mock jwt claims
			ts.App.Use(func(c *fiber.Ctx) error {
				c.Locals("user", &jwtx.Claims{UserID: test.userID, Username: "Test", Role: "cashier"})
				return c.Next()
			})

			// Create Route
			ts.App.Delete("/user/:id", ts.Handler.DeleteProfile)
			test.setup(ts)

			// create request
			req := httptest.NewRequest(fiber.MethodDelete, test.path, nil)
			req.Header.Set(fiber.HeaderAuthorization, test.haederValue)

			// run test
			res, _ := ts.App.Test(req, -1)

			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			if test.expectedBody != nil {
				expectedBody, _ := json.Marshal(test.expectedBody)
				resBody, _ := io.ReadAll(res.Body)

				assert.JSONEq(t, string(expectedBody), string(resBody))
			}
		})
	}
}
