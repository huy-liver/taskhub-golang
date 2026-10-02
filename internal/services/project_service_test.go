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

func TestProjectService_Create(t *testing.T) {
	mockRepo := new(MockProjectRepository)
	ownerID := uint(1)

	req := dto.CreateProjectRequest{
		Name:        "Backend Redesign",
		Description: "Refactoring the entire backend",
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Project) bool {
		return p.Name == req.Name && p.OwnerID == ownerID
	})).Return(nil)

	mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(&models.Project{
		ID:          1,
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	svc := services.NewProjectService(mockRepo)
	res, appErr := svc.Create(context.Background(), ownerID, &req)

	require.Nil(t, appErr)
	require.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, ownerID, res.OwnerID)
	mockRepo.AssertExpectations(t)
}

func TestProjectService_Update(t *testing.T) {
	ownerID := uint(1)
	strangerID := uint(2)

	existingProject := &models.Project{
		ID:          10,
		Name:        "Original Name",
		Description: "Original Desc",
		OwnerID:     ownerID,
	}

	tests := []struct {
		name          string
		id            uint
		userID        uint
		role          string
		req           dto.UpdateProjectRequest
		mockSetup     func(mockRepo *MockProjectRepository)
		expectedError string
	}{
		{
			name:   "Success - Owner updates project",
			id:     10,
			userID: ownerID,
			role:   "user",
			req: dto.UpdateProjectRequest{
				Name: "Updated Name",
			},
			mockSetup: func(mockRepo *MockProjectRepository) {
				mockRepo.On("GetByID", mock.Anything, uint(10)).Return(existingProject, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("CountTasks", mock.Anything, uint(10)).Return(3, nil)
			},
			expectedError: "",
		},
		{
			name:   "Forbidden - Non-owner user updates project",
			id:     10,
			userID: strangerID,
			role:   "user",
			req: dto.UpdateProjectRequest{
				Name: "Hacked Name",
			},
			mockSetup: func(mockRepo *MockProjectRepository) {
				mockRepo.On("GetByID", mock.Anything, uint(10)).Return(existingProject, nil)
			},
			expectedError: "You do not have permission to update this project",
		},
		{
			name:   "Success - Admin updates any project",
			id:     10,
			userID: strangerID,
			role:   "admin",
			req: dto.UpdateProjectRequest{
				Name: "Admin Update",
			},
			mockSetup: func(mockRepo *MockProjectRepository) {
				mockRepo.On("GetByID", mock.Anything, uint(10)).Return(existingProject, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				mockRepo.On("CountTasks", mock.Anything, uint(10)).Return(3, nil)
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockProjectRepository)
			tc.mockSetup(mockRepo)

			svc := services.NewProjectService(mockRepo)
			res, appErr := svc.Update(context.Background(), tc.id, tc.userID, tc.role, &tc.req)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
				assert.Nil(t, res)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, res)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_Delete(t *testing.T) {
	ownerID := uint(1)
	strangerID := uint(2)

	existingProject := &models.Project{
		ID:      10,
		OwnerID: ownerID,
	}

	tests := []struct {
		name          string
		id            uint
		userID        uint
		role          string
		mockSetup     func(mockRepo *MockProjectRepository)
		expectedError string
	}{
		{
			name:   "Success - Owner deletes project",
			id:     10,
			userID: ownerID,
			role:   "user",
			mockSetup: func(mockRepo *MockProjectRepository) {
				mockRepo.On("GetByID", mock.Anything, uint(10)).Return(existingProject, nil)
				mockRepo.On("Delete", mock.Anything, uint(10)).Return(nil)
			},
			expectedError: "",
		},
		{
			name:   "Forbidden - Non-owner deletes project",
			id:     10,
			userID: strangerID,
			role:   "user",
			mockSetup: func(mockRepo *MockProjectRepository) {
				mockRepo.On("GetByID", mock.Anything, uint(10)).Return(existingProject, nil)
			},
			expectedError: "You do not have permission to delete this project",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockProjectRepository)
			tc.mockSetup(mockRepo)

			svc := services.NewProjectService(mockRepo)
			appErr := svc.Delete(context.Background(), tc.id, tc.userID, tc.role)

			if tc.expectedError != "" {
				require.NotNil(t, appErr)
				assert.Equal(t, tc.expectedError, appErr.Message)
			} else {
				require.Nil(t, appErr)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_GetByID(t *testing.T) {
	mockRepo := new(MockProjectRepository)
	project := &models.Project{
		ID:   1,
		Name: "Test Project",
	}

	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(project, nil)
	mockRepo.On("CountTasks", mock.Anything, uint(1)).Return(5, nil)
	mockRepo.On("GetByID", mock.Anything, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	svc := services.NewProjectService(mockRepo)
	res, appErr := svc.GetByID(context.Background(), 1)
	require.Nil(t, appErr)
	require.NotNil(t, res)
	assert.Equal(t, uint(1), res.ID)
	assert.Equal(t, 5, res.TasksCount)

	_, notFoundErr := svc.GetByID(context.Background(), 99)
	require.NotNil(t, notFoundErr)
	assert.Equal(t, "Project not found", notFoundErr.Message)

	mockRepo.AssertExpectations(t)
}

func TestProjectService_List(t *testing.T) {
	mockRepo := new(MockProjectRepository)
	pagination := utils.PaginationQuery{Page: 1, Limit: 10}

	projects := []models.Project{
		{ID: 1, Name: "Project 1"},
		{ID: 2, Name: "Project 2"},
	}

	mockRepo.On("List", mock.Anything, pagination, (*uint)(nil)).Return(projects, int64(2), nil)
	mockRepo.On("CountTasks", mock.Anything, uint(1)).Return(2, nil)
	mockRepo.On("CountTasks", mock.Anything, uint(2)).Return(4, nil)

	svc := services.NewProjectService(mockRepo)
	res, total, appErr := svc.List(context.Background(), pagination, nil)

	require.Nil(t, appErr)
	assert.Equal(t, int64(2), total)
	assert.Len(t, res, 2)
	mockRepo.AssertExpectations(t)
}
