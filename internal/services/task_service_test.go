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

func TestTaskService_Create(t *testing.T) {
	projectOwnerID := uint(1)
	otherUserID := uint(2)
	assigneeID := uint(3)

	project := &models.Project{
		ID:      10,
		Name:    "Project Alpha",
		OwnerID: projectOwnerID,
	}

	tests := []struct {
		name          string
		userID        uint
		role          string
		req           dto.CreateTaskRequest
		mockSetup     func(tRepo *MockTaskRepository, pRepo *MockProjectRepository, uRepo *MockUserRepository)
		expectedError string
	}{
		{
			name:   "Success - Created by project owner",
			userID: projectOwnerID,
			role:   "user",
			req: dto.CreateTaskRequest{
				Title:      "Implement Login",
				ProjectID:  10,
				AssigneeID: &assigneeID,
				Status:     "todo",
				Priority:   "high",
			},
			mockSetup: func(tRepo *MockTaskRepository, pRepo *MockProjectRepository, uRepo *MockUserRepository) {
				pRepo.On("GetByID", mock.Anything, uint(10)).Return(project, nil)
				uRepo.On("GetByID", mock.Anything, assigneeID).Return(&models.User{ID: assigneeID}, nil)
				tRepo.On("Create", mock.Anything, mock.MatchedBy(func(t *models.Task) bool {
					return t.Title == "Implement Login" && t.ProjectID == 10
				})).Return(nil)
				tRepo.On("GetByID", mock.Anything, mock.Anything).Return(&models.Task{
					ID:        1,
					Title:     "Implement Login",
					ProjectID: 10,
					Status:    "todo",
					Priority:  "high",
				}, nil)
			},
			expectedError: "",
		},
		{
			name:   "Forbidden - Non-owner user creates task",
			userID: otherUserID,
			role:   "user",
			req: dto.CreateTaskRequest{
				Title:     "Unauthorized Task",
				ProjectID: 10,
			},
			mockSetup: func(tRepo *MockTaskRepository, pRepo *MockProjectRepository, uRepo *MockUserRepository) {
				pRepo.On("GetByID", mock.Anything, uint(10)).Return(project, nil)
			},
			expectedError: "Only project owner or admin can add tasks to this project",
		},
		{
			name:   "Not Found - Project does not exist",
			userID: 1,
			role:   "admin",
			req: dto.CreateTaskRequest{
				Title:     "Non-existent Project Task",
				ProjectID: 999,
			},
			mockSetup: func(tRepo *MockTaskRepository, pRepo *MockProjectRepository, uRepo *MockUserRepository) {
				pRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "Project not found",
		},
		{
			name:   "Bad Request - Assignee user does not exist",
			userID: projectOwnerID,
			role:   "user",
			req: dto.CreateTaskRequest{
				Title:      "Missing Assignee Task",
				ProjectID:  10,
				AssigneeID: &otherUserID,
			},
			mockSetup: func(tRepo *MockTaskRepository, pRepo *MockProjectRepository, uRepo *MockUserRepository) {
				pRepo.On("GetByID", mock.Anything, uint(10)).Return(project, nil)
				uRepo.On("GetByID", mock.Anything, otherUserID).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "Assignee user does not exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tRepo := new(MockTaskRepository)
			pRepo := new(MockProjectRepository)
			uRepo := new(MockUserRepository)
			tc.mockSetup(tRepo, pRepo, uRepo)

			taskSvc := services.NewTaskService(tRepo, pRepo, uRepo)
			res, appErr := taskSvc.Create(context.Background(), tc.userID, tc.role, &tc.req)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
				assert.Nil(t, res)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, res)
				assert.Equal(t, tc.req.Title, res.Title)
			}

			tRepo.AssertExpectations(t)
			pRepo.AssertExpectations(t)
			uRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_GetByID(t *testing.T) {
	task := &models.Task{
		ID:        5,
		Title:     "Write Unit Tests",
		Status:    "in_progress",
		Priority:  "medium",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		id            uint
		mockSetup     func(tRepo *MockTaskRepository)
		expectedError string
	}{
		{
			name: "Success - Task found",
			id:   5,
			mockSetup: func(tRepo *MockTaskRepository) {
				tRepo.On("GetByID", mock.Anything, uint(5)).Return(task, nil)
			},
			expectedError: "",
		},
		{
			name: "Not Found - Task does not exist",
			id:   99,
			mockSetup: func(tRepo *MockTaskRepository) {
				tRepo.On("GetByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: "Task not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tRepo := new(MockTaskRepository)
			pRepo := new(MockProjectRepository)
			uRepo := new(MockUserRepository)
			tc.mockSetup(tRepo)

			taskSvc := services.NewTaskService(tRepo, pRepo, uRepo)
			res, appErr := taskSvc.GetByID(context.Background(), tc.id)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
				assert.Nil(t, res)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, res)
				assert.Equal(t, tc.id, res.ID)
			}

			tRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_PatchStatus(t *testing.T) {
	ownerID := uint(1)
	assigneeID := uint(2)
	strangerID := uint(99)

	existingTask := &models.Task{
		ID:         1,
		Title:      "Refactor DB",
		Status:     "todo",
		AssigneeID: &assigneeID,
		Project: &models.Project{
			ID:      10,
			OwnerID: ownerID,
		},
	}

	tests := []struct {
		name          string
		id            uint
		userID        uint
		role          string
		newStatus     string
		mockSetup     func(tRepo *MockTaskRepository)
		expectedError string
	}{
		{
			name:      "Success - Assignee updates status",
			id:        1,
			userID:    assigneeID,
			role:      "user",
			newStatus: "in_progress",
			mockSetup: func(tRepo *MockTaskRepository) {
				tRepo.On("GetByID", mock.Anything, uint(1)).Return(existingTask, nil).Once()
				tRepo.On("PatchStatus", mock.Anything, uint(1), "in_progress").Return(nil)
				tRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.Task{
					ID:         1,
					Title:      "Refactor DB",
					Status:     "in_progress",
					AssigneeID: &assigneeID,
				}, nil).Once()
			},
			expectedError: "",
		},
		{
			name:      "Forbidden - Unrelated user updates status",
			id:        1,
			userID:    strangerID,
			role:      "user",
			newStatus: "done",
			mockSetup: func(tRepo *MockTaskRepository) {
				tRepo.On("GetByID", mock.Anything, uint(1)).Return(existingTask, nil)
			},
			expectedError: "You do not have permission to update this task status",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tRepo := new(MockTaskRepository)
			pRepo := new(MockProjectRepository)
			uRepo := new(MockUserRepository)
			tc.mockSetup(tRepo)

			taskSvc := services.NewTaskService(tRepo, pRepo, uRepo)
			res, appErr := taskSvc.PatchStatus(context.Background(), tc.id, tc.userID, tc.role, tc.newStatus)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
				assert.Nil(t, res)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, res)
				assert.Equal(t, tc.newStatus, res.Status)
			}

			tRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_Delete(t *testing.T) {
	ownerID := uint(1)
	strangerID := uint(2)

	task := &models.Task{
		ID: 1,
		Project: &models.Project{
			ID:      10,
			OwnerID: ownerID,
		},
	}

	tests := []struct {
		name          string
		id            uint
		userID        uint
		role          string
		mockSetup     func(tRepo *MockTaskRepository)
		expectedError string
	}{
		{
			name:   "Success - Owner deletes task",
			id:     1,
			userID: ownerID,
			role:   "user",
			mockSetup: func(tRepo *MockTaskRepository) {
				tRepo.On("GetByID", mock.Anything, uint(1)).Return(task, nil)
				tRepo.On("Delete", mock.Anything, uint(1)).Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "Forbidden - Non-owner deletes task",
			id:     1,
			userID: strangerID,
			role:   "user",
			mockSetup: func(tRepo *MockTaskRepository) {
				tRepo.On("GetByID", mock.Anything, uint(1)).Return(task, nil)
			},
			expectedError: "Only project owner or admin can delete this task",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tRepo := new(MockTaskRepository)
			pRepo := new(MockProjectRepository)
			uRepo := new(MockUserRepository)
			tc.mockSetup(tRepo)

			taskSvc := services.NewTaskService(tRepo, pRepo, uRepo)
			appErr := taskSvc.Delete(context.Background(), tc.id, tc.userID, tc.role)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
			} else {
				require.Nil(t, appErr)
			}

			tRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_List(t *testing.T) {
	tRepo := new(MockTaskRepository)
	pRepo := new(MockProjectRepository)
	uRepo := new(MockUserRepository)

	pagination := utils.PaginationQuery{Page: 1, Limit: 10}
	filter := dto.TaskFilterQuery{Status: "todo"}

	tasks := []models.Task{
		{ID: 1, Title: "Task 1", Status: "todo"},
		{ID: 2, Title: "Task 2", Status: "todo"},
	}

	tRepo.On("List", mock.Anything, pagination, filter).Return(tasks, int64(2), nil)

	taskSvc := services.NewTaskService(tRepo, pRepo, uRepo)
	res, total, appErr := taskSvc.List(context.Background(), pagination, filter)

	require.Nil(t, appErr)
	assert.Equal(t, int64(2), total)
	assert.Len(t, res, 2)
	tRepo.AssertExpectations(t)
}

func TestTaskService_Update(t *testing.T) {
	ownerID := uint(1)
	strangerID := uint(99)

	existingTask := &models.Task{
		ID:    1,
		Title: "Old Task Title",
		Project: &models.Project{
			ID:      10,
			OwnerID: ownerID,
		},
	}

	t.Run("Success - Owner updates task", func(t *testing.T) {
		tRepo := new(MockTaskRepository)
		pRepo := new(MockProjectRepository)
		uRepo := new(MockUserRepository)

		tRepo.On("GetByID", mock.Anything, uint(1)).Return(existingTask, nil)
		tRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		taskSvc := services.NewTaskService(tRepo, pRepo, uRepo)
		res, appErr := taskSvc.Update(context.Background(), 1, ownerID, "user", &dto.UpdateTaskRequest{
			Title: "New Task Title",
		})

		require.Nil(t, appErr)
		require.NotNil(t, res)
		tRepo.AssertExpectations(t)
	})

	t.Run("Forbidden - Stranger cannot update task", func(t *testing.T) {
		tRepo := new(MockTaskRepository)
		pRepo := new(MockProjectRepository)
		uRepo := new(MockUserRepository)

		tRepo.On("GetByID", mock.Anything, uint(1)).Return(existingTask, nil)

		taskSvc := services.NewTaskService(tRepo, pRepo, uRepo)
		res, appErr := taskSvc.Update(context.Background(), 1, strangerID, "user", &dto.UpdateTaskRequest{
			Title: "Hacked Title",
		})

		require.NotNil(t, appErr)
		assert.Nil(t, res)
		assert.Equal(t, "You do not have permission to update this task", appErr.Message)
		tRepo.AssertExpectations(t)
	})
}
