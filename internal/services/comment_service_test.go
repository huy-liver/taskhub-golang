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

func TestCommentService_CRUD(t *testing.T) {
	cRepo := new(MockCommentRepository)
	tRepo := new(MockTaskRepository)
	svc := services.NewCommentService(cRepo, tRepo)

	t.Run("Create Comment - Success", func(t *testing.T) {
		req := &dto.CreateCommentRequest{
			TaskID:  1,
			Content: "This looks great!",
		}

		tRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.Task{ID: 1}, nil).Once()
		cRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		cRepo.On("GetByID", mock.Anything, mock.Anything).Return(&models.Comment{
			ID:        1,
			Content:   req.Content,
			TaskID:    1,
			UserID:    10,
			CreatedAt: time.Now(),
		}, nil).Once()

		res, appErr := svc.Create(context.Background(), 10, req)
		require.Nil(t, appErr)
		assert.Equal(t, req.Content, res.Content)
	})

	t.Run("Create Comment - Task Not Found", func(t *testing.T) {
		req := &dto.CreateCommentRequest{
			TaskID:  99,
			Content: "Task missing",
		}
		tRepo.On("GetByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound).Once()

		res, appErr := svc.Create(context.Background(), 10, req)
		require.NotNil(t, appErr)
		assert.Nil(t, res)
	})

	t.Run("GetByID Comment - Success", func(t *testing.T) {
		comment := &models.Comment{
			ID:      1,
			Content: "Sample comment",
			UserID:  10,
		}
		cRepo.On("GetByID", mock.Anything, uint(1)).Return(comment, nil).Once()

		res, appErr := svc.GetByID(context.Background(), 1)
		require.Nil(t, appErr)
		assert.Equal(t, uint(1), res.ID)
		assert.Equal(t, "Sample comment", res.Content)
	})

	t.Run("GetByID Comment - Not Found", func(t *testing.T) {
		cRepo.On("GetByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound).Once()

		res, appErr := svc.GetByID(context.Background(), 99)
		require.NotNil(t, appErr)
		assert.Nil(t, res)
		assert.Equal(t, "Comment not found", appErr.Message)
	})

	t.Run("Update Comment - Success by Author", func(t *testing.T) {
		comment := &models.Comment{
			ID:      1,
			Content: "Old comment",
			UserID:  10,
		}
		cRepo.On("GetByID", mock.Anything, uint(1)).Return(comment, nil).Once()
		cRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

		res, appErr := svc.Update(context.Background(), 1, 10, "user", &dto.UpdateCommentRequest{Content: "Updated comment"})
		require.Nil(t, appErr)
		assert.Equal(t, "Updated comment", res.Content)
	})

	t.Run("Update Comment - Forbidden by Non-Author", func(t *testing.T) {
		comment := &models.Comment{
			ID:      1,
			Content: "Old comment",
			UserID:  10,
		}
		cRepo.On("GetByID", mock.Anything, uint(1)).Return(comment, nil).Once()

		res, appErr := svc.Update(context.Background(), 1, 99, "user", &dto.UpdateCommentRequest{Content: "Hacked comment"})
		require.NotNil(t, appErr)
		assert.Equal(t, "You do not have permission to update this comment", appErr.Message)
		assert.Nil(t, res)
	})

	t.Run("Delete Comment - Success by Author", func(t *testing.T) {
		comment := &models.Comment{
			ID:     1,
			UserID: 10,
		}
		cRepo.On("GetByID", mock.Anything, uint(1)).Return(comment, nil).Once()
		cRepo.On("Delete", mock.Anything, uint(1)).Return(nil).Once()

		appErr := svc.Delete(context.Background(), 1, 10, "user")
		require.Nil(t, appErr)
	})

	t.Run("List Comments By Task", func(t *testing.T) {
		pagination := utils.PaginationQuery{Page: 1, Limit: 10}
		comments := []models.Comment{
			{ID: 1, Content: "Comment 1", TaskID: 1},
		}
		cRepo.On("ListByTask", mock.Anything, uint(1), pagination).Return(comments, int64(1), nil).Once()

		res, total, appErr := svc.ListByTask(context.Background(), 1, pagination)
		require.Nil(t, appErr)
		assert.Equal(t, int64(1), total)
		assert.Len(t, res, 1)
	})
}
