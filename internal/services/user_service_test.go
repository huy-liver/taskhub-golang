package services_test

import (
	"context"
	"testing"
	"time"

	"taskhub/internal/dto"
	"taskhub/internal/models"
	"taskhub/internal/services"
	"taskhub/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserService_CRUD(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	t.Run("Create User", func(t *testing.T) {
		req := &dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@test.com",
			Password: "password123",
		}

		mockRepo.On("GetByEmail", mock.Anything, "test@test.com").Return(nil, gorm.ErrRecordNotFound).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		res, appErr := svc.Create(context.Background(), req)
		require.Nil(t, appErr)
		assert.Equal(t, req.Email, res.Email)
	})

	t.Run("GetByID User - Found", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@test.com",
		}
		mockRepo.On("GetByID", mock.Anything, uint(1)).Return(user, nil).Once()

		res, appErr := svc.GetByID(context.Background(), 1)
		require.Nil(t, appErr)
		assert.Equal(t, uint(1), res.ID)
	})

	t.Run("GetByID User - Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound).Once()

		res, appErr := svc.GetByID(context.Background(), 99)
		require.NotNil(t, appErr)
		assert.Nil(t, res)
	})

	t.Run("Update User", func(t *testing.T) {
		user := &models.User{
			ID:   1,
			Name: "Old Name",
		}
		mockRepo.On("GetByID", mock.Anything, uint(1)).Return(user, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		res, appErr := svc.Update(context.Background(), 1, &dto.UpdateUserRequest{Name: "New Name"})
		require.Nil(t, appErr)
		assert.Equal(t, "New Name", res.Name)
	})

	t.Run("Delete User", func(t *testing.T) {
		user := &models.User{ID: 1}
		mockRepo.On("GetByID", mock.Anything, uint(1)).Return(user, nil).Once()
		mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil).Once()

		appErr := svc.Delete(context.Background(), 1)
		require.Nil(t, appErr)
	})

	t.Run("List Users", func(t *testing.T) {
		pagination := utils.PaginationQuery{Page: 1, Limit: 10}
		users := []models.User{
			{ID: 1, Name: "User 1", CreatedAt: time.Now()},
		}
		mockRepo.On("List", mock.Anything, pagination).Return(users, int64(1), nil).Once()

		res, total, appErr := svc.List(context.Background(), pagination)
		require.Nil(t, appErr)
		assert.Equal(t, int64(1), total)
		assert.Len(t, res, 1)
	})
}
