package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"taskhub/internal/dto"
	"taskhub/internal/errors"
	"taskhub/internal/handlers"
	"taskhub/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService mocks the AuthService interface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, *errors.AppError) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.AppError)
	}
	return args.Get(0).(*dto.AuthResponse), nil
}

func (m *MockAuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, *errors.AppError) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.AppError)
	}
	return args.Get(0).(*dto.AuthResponse), nil
}

func (m *MockAuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, *errors.AppError) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.AppError)
	}
	return args.Get(0).(*dto.AuthResponse), nil
}

func (m *MockAuthService) GetMe(ctx context.Context, userID uint) (*dto.UserResponse, *errors.AppError) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.AppError)
	}
	return args.Get(0).(*dto.UserResponse), nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthHandler_Register(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := handlers.NewAuthHandler(mockSvc)

	r := setupTestRouter()
	r.POST("/api/v1/auth/register", handler.Register)

	t.Run("Success - 201 Created", func(t *testing.T) {
		reqBody := dto.RegisterRequest{
			Name:     "Alice Smith",
			Email:    "alice@example.com",
			Password: "securepassword",
		}
		jsonBytes, _ := json.Marshal(reqBody)

		mockSvc.On("Register", mock.Anything, &reqBody).Return(&dto.AuthResponse{
			User: dto.UserResponse{
				ID:    1,
				Name:  reqBody.Name,
				Email: reqBody.Email,
			},
			AccessToken: "valid_access_token",
		}, nil).Once()

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp utils.StandardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("Validation Error - Missing required fields", func(t *testing.T) {
		reqBody := map[string]string{
			"email": "not-an-email",
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := handlers.NewAuthHandler(mockSvc)

	r := setupTestRouter()
	r.POST("/api/v1/auth/login", handler.Login)

	t.Run("Success - 200 OK", func(t *testing.T) {
		reqBody := dto.LoginRequest{
			Email:    "alice@example.com",
			Password: "securepassword",
		}
		jsonBytes, _ := json.Marshal(reqBody)

		mockSvc.On("Login", mock.Anything, &reqBody).Return(&dto.AuthResponse{
			AccessToken: "access_token_123",
		}, nil).Once()

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unauthorized - 401 on wrong password", func(t *testing.T) {
		reqBody := dto.LoginRequest{
			Email:    "alice@example.com",
			Password: "wrongpassword",
		}
		jsonBytes, _ := json.Marshal(reqBody)

		mockSvc.On("Login", mock.Anything, &reqBody).Return(nil, errors.NewUnauthorizedError("Invalid email or password")).Once()

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
