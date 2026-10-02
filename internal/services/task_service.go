package services

import (
	"context"
	"taskhub/internal/dto"
	"taskhub/internal/errors"
	"taskhub/internal/models"
	"taskhub/internal/repositories"
	"taskhub/internal/utils"

	"gorm.io/gorm"
)

// TaskService defines interface for task business operations
type TaskService interface {
	Create(ctx context.Context, userID uint, role string, req *dto.CreateTaskRequest) (*dto.TaskResponse, *errors.AppError)
	GetByID(ctx context.Context, id uint) (*dto.TaskResponse, *errors.AppError)
	Update(ctx context.Context, id uint, userID uint, role string, req *dto.UpdateTaskRequest) (*dto.TaskResponse, *errors.AppError)
	PatchStatus(ctx context.Context, id uint, userID uint, role string, status string) (*dto.TaskResponse, *errors.AppError)
	Delete(ctx context.Context, id uint, userID uint, role string) *errors.AppError
	List(ctx context.Context, pagination utils.PaginationQuery, filter dto.TaskFilterQuery) ([]dto.TaskResponse, int64, *errors.AppError)
}

type taskService struct {
	taskRepo    repositories.TaskRepository
	projectRepo repositories.ProjectRepository
	userRepo    repositories.UserRepository
}

// NewTaskService creates a new TaskService instance
func NewTaskService(taskRepo repositories.TaskRepository, projectRepo repositories.ProjectRepository, userRepo repositories.UserRepository) TaskService {
	return &taskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

func (s *taskService) Create(ctx context.Context, userID uint, role string, req *dto.CreateTaskRequest) (*dto.TaskResponse, *errors.AppError) {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Project not found")
		}
		return nil, errors.NewInternalServerError("Failed to check project", err)
	}

	// Verify authorization: only project owner or admin can add tasks to the project
	if role != models.RoleAdmin && project.OwnerID != userID {
		return nil, errors.NewForbiddenError("Only project owner or admin can add tasks to this project")
	}

	// Verify assignee exists if specified
	if req.AssigneeID != nil {
		_, err := s.userRepo.GetByID(ctx, *req.AssigneeID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.NewBadRequestError("Assignee user does not exist", nil)
			}
			return nil, errors.NewInternalServerError("Failed to verify assignee", err)
		}
	}

	status := models.TaskStatusTodo
	if req.Status != "" {
		status = req.Status
	}

	priority := models.TaskPriorityMedium
	if req.Priority != "" {
		priority = req.Priority
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     req.DueDate,
		ProjectID:   req.ProjectID,
		AssigneeID:  req.AssigneeID,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, errors.NewInternalServerError("Failed to create task", err)
	}

	created, err := s.taskRepo.GetByID(ctx, task.ID)
	if err != nil {
		return toTaskResponse(task), nil
	}

	return toTaskResponse(created), nil
}

func (s *taskService) GetByID(ctx context.Context, id uint) (*dto.TaskResponse, *errors.AppError) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Task not found")
		}
		return nil, errors.NewInternalServerError("Failed to get task", err)
	}
	return toTaskResponse(task), nil
}

func (s *taskService) Update(ctx context.Context, id uint, userID uint, role string, req *dto.UpdateTaskRequest) (*dto.TaskResponse, *errors.AppError) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Task not found")
		}
		return nil, errors.NewInternalServerError("Failed to get task", err)
	}

	// Authorization check: project owner, assignee, or admin can update
	isOwner := task.Project != nil && task.Project.OwnerID == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	if role != models.RoleAdmin && !isOwner && !isAssignee {
		return nil, errors.NewForbiddenError("You do not have permission to update this task")
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.AssigneeID != nil {
		// Only project owner or admin can change assignee
		if role != models.RoleAdmin && !isOwner {
			return nil, errors.NewForbiddenError("Only project owner or admin can change assignee")
		}
		// Verify new assignee exists
		_, err := s.userRepo.GetByID(ctx, *req.AssigneeID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.NewBadRequestError("Assignee user does not exist", nil)
			}
			return nil, errors.NewInternalServerError("Failed to verify assignee", err)
		}
		task.AssigneeID = req.AssigneeID
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, errors.NewInternalServerError("Failed to update task", err)
	}

	updated, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return toTaskResponse(task), nil
	}
	return toTaskResponse(updated), nil
}

func (s *taskService) PatchStatus(ctx context.Context, id uint, userID uint, role string, status string) (*dto.TaskResponse, *errors.AppError) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Task not found")
		}
		return nil, errors.NewInternalServerError("Failed to get task", err)
	}

	// Owner, assignee, or admin can update status
	isOwner := task.Project != nil && task.Project.OwnerID == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	if role != models.RoleAdmin && !isOwner && !isAssignee {
		return nil, errors.NewForbiddenError("You do not have permission to update this task status")
	}

	if err := s.taskRepo.PatchStatus(ctx, id, status); err != nil {
		return nil, errors.NewInternalServerError("Failed to patch task status", err)
	}

	updated, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		task.Status = status
		return toTaskResponse(task), nil
	}
	return toTaskResponse(updated), nil
}

func (s *taskService) Delete(ctx context.Context, id uint, userID uint, role string) *errors.AppError {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("Task not found")
		}
		return errors.NewInternalServerError("Failed to get task", err)
	}

	// Only project owner or admin can delete task
	isOwner := task.Project != nil && task.Project.OwnerID == userID
	if role != models.RoleAdmin && !isOwner {
		return errors.NewForbiddenError("Only project owner or admin can delete this task")
	}

	if err := s.taskRepo.Delete(ctx, id); err != nil {
		return errors.NewInternalServerError("Failed to delete task", err)
	}
	return nil
}

func (s *taskService) List(ctx context.Context, pagination utils.PaginationQuery, filter dto.TaskFilterQuery) ([]dto.TaskResponse, int64, *errors.AppError) {
	tasks, total, err := s.taskRepo.List(ctx, pagination, filter)
	if err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve tasks", err)
	}

	var responses []dto.TaskResponse
	for _, t := range tasks {
		responses = append(responses, *toTaskResponse(&t))
	}
	return responses, total, nil
}

func toTaskResponse(t *models.Task) *dto.TaskResponse {
	if t == nil {
		return nil
	}
	return &dto.TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		DueDate:     t.DueDate,
		ProjectID:   t.ProjectID,
		Project:     toProjectResponse(t.Project, 0),
		AssigneeID:  t.AssigneeID,
		Assignee:    toUserResponse(t.Assignee),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
