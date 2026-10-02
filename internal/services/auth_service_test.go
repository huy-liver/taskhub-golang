package services_test

import (
	"context"
	"testing"
	"time"

	"taskhub/internal/configs"
	"taskhub/internal/dto"
	"taskhub/internal/models"
	"taskhub/internal/services"
	"taskhub/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func getTestConfig() *configs.Config {
	return &configs.Config{
		JWTSecret:           "test_jwt_secret_key_123456789012345",
		JWTExpiresIn:        "1h",
		JWTRefreshSecret:    "test_jwt_refresh_secret_key_1234567",
		JWTRefreshExpiresIn: "24h",
	}
}

func TestAuthService_Register(t *testing.T) {
	cfg := getTestConfig()

	tests := []struct {
		name          string
		req           dto.RegisterRequest
		mockSetup     func(mockRepo *MockUserRepository)
		expectedError string
	}{
		{
			name: "Success - Register new user",
			req: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     "user",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "john@example.com").Return(nil, gorm.ErrRecordNotFound)
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
					return u.Email == "john@example.com" && u.Name == "John Doe"
				})).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "Conflict - Email already exists",
			req: dto.RegisterRequest{
				Name:     "Existing User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(&models.User{
					ID:    1,
					Email: "existing@example.com",
				}, nil)
			},
			expectedError: "Email is already registered",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			authSvc := services.NewAuthService(mockRepo, cfg)
			res, appErr := authSvc.Register(context.Background(), &tc.req)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
				assert.Nil(t, res)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, res)
				assert.Equal(t, tc.req.Email, res.User.Email)
				assert.NotEmpty(t, res.AccessToken)
				assert.NotEmpty(t, res.RefreshToken)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	cfg := getTestConfig()
	hashedPassword, err := utils.HashPassword("secret123")
	require.NoError(t, err)

	existingUser := &models.User{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Password:  hashedPassword,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		req           dto.LoginRequest
		mockSetup     func(mockRepo *MockUserRepository)
		expectedError string
	}{
		{
			name: "Success - Valid credentials",
			req: dto.LoginRequest{
				Email:    "alice@example.com",
				Password: "secret123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "alice@example.com").Return(existingUser, nil)
			},
			expectedError: "",
		},
		{
			name: "Unauthorized - User not found",
			req: dto.LoginRequest{
				Email:    "notfound@example.com",
				Password: "secret123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "Invalid email or password",
		},
		{
			name: "Unauthorized - Incorrect password",
			req: dto.LoginRequest{
				Email:    "alice@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "alice@example.com").Return(existingUser, nil)
			},
			expectedError: "Invalid email or password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			authSvc := services.NewAuthService(mockRepo, cfg)
			res, appErr := authSvc.Login(context.Background(), &tc.req)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
				assert.Nil(t, res)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, res)
				assert.Equal(t, tc.req.Email, res.User.Email)
				assert.NotEmpty(t, res.AccessToken)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetMe(t *testing.T) {
	cfg := getTestConfig()

	user := &models.User{
		ID:        10,
		Name:      "Bob",
		Email:     "bob@example.com",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		userID        uint
		mockSetup     func(mockRepo *MockUserRepository)
		expectedError string
	}{
		{
			name:   "Success - Found user",
			userID: 10,
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, uint(10)).Return(user, nil)
			},
			expectedError: "",
		},
		{
			name:   "Not Found - User not found",
			userID: 99,
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "User not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tc.mockSetup(mockRepo)

			authSvc := services.NewAuthService(mockRepo, cfg)
			res, appErr := authSvc.GetMe(context.Background(), tc.userID)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
				assert.Nil(t, res)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, res)
				assert.Equal(t, tc.userID, res.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	cfg := getTestConfig()
	validRefreshToken, _ := utils.GenerateToken(5, "user", cfg.JWTRefreshSecret, time.Hour)

	user := &models.User{
		ID:        5,
		Name:      "Eve",
		Email:     "eve@example.com",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("Success - Valid refresh token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByID", mock.Anything, uint(5)).Return(user, nil)

		authSvc := services.NewAuthService(mockRepo, cfg)
		res, appErr := authSvc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
			RefreshToken: validRefreshToken,
		})

		require.Nil(t, appErr)
		require.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.Equal(t, uint(5), res.User.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Unauthorized - Invalid refresh token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		authSvc := services.NewAuthService(mockRepo, cfg)
		res, appErr := authSvc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
			RefreshToken: "invalid_refresh_token",
		})

		require.NotNil(t, appErr)
		assert.Nil(t, res)
		assert.Equal(t, "Invalid or expired refresh token", appErr.Message)
	})
}
