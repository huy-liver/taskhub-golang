package services

import (
	"context"
	"fmt"
	"taskhub/internal/cache"
	"taskhub/internal/dto"
	"taskhub/internal/errors"
	"taskhub/internal/models"
	"taskhub/internal/repositories"
	"taskhub/internal/utils"
	"time"

	"gorm.io/gorm"
)

// ProjectService defines interface for project business operations
type ProjectService interface {
	Create(ctx context.Context, ownerID uint, req *dto.CreateProjectRequest) (*dto.ProjectResponse, *errors.AppError)
	GetByID(ctx context.Context, id uint) (*dto.ProjectResponse, *errors.AppError)
	Update(ctx context.Context, id uint, userID uint, role string, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, *errors.AppError)
	Delete(ctx context.Context, id uint, userID uint, role string) *errors.AppError
	List(ctx context.Context, pagination utils.PaginationQuery, ownerID *uint) ([]dto.ProjectResponse, int64, *errors.AppError)
}

type projectService struct {
	projectRepo repositories.ProjectRepository
	cache       *cache.RedisClient
	cacheTTL    time.Duration
}

// NewProjectService creates a new ProjectService instance with optional Redis cache
func NewProjectService(projectRepo repositories.ProjectRepository, cacheClient ...*cache.RedisClient) ProjectService {
	var c *cache.RedisClient
	if len(cacheClient) > 0 && cacheClient[0] != nil {
		c = cacheClient[0]
	}
	return &projectService{
		projectRepo: projectRepo,
		cache:       c,
		cacheTTL:    15 * time.Minute,
	}
}

func (s *projectService) Create(ctx context.Context, ownerID uint, req *dto.CreateProjectRequest) (*dto.ProjectResponse, *errors.AppError) {
	project := &models.Project{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, errors.NewInternalServerError("Failed to create project", err)
	}

	s.invalidateCache(ctx)

	// Fetch with owner relation
	created, err := s.projectRepo.GetByID(ctx, project.ID)
	if err != nil {
		return toProjectResponse(project, 0), nil
	}

	return toProjectResponse(created, 0), nil
}

func (s *projectService) GetByID(ctx context.Context, id uint) (*dto.ProjectResponse, *errors.AppError) {
	cacheKey := fmt.Sprintf("project:detail:%d", id)
	if s.cache != nil {
		var cached dto.ProjectResponse
		if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Project not found")
		}
		return nil, errors.NewInternalServerError("Failed to get project", err)
	}

	taskCount, _ := s.projectRepo.CountTasks(ctx, id)
	res := toProjectResponse(project, taskCount)

	if s.cache != nil && res != nil {
		_ = s.cache.Set(ctx, cacheKey, res, s.cacheTTL)
	}

	return res, nil
}

func (s *projectService) Update(ctx context.Context, id uint, userID uint, role string, req *dto.UpdateProjectRequest) (*dto.ProjectResponse, *errors.AppError) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Project not found")
		}
		return nil, errors.NewInternalServerError("Failed to get project", err)
	}

	// Authorization check: Only project owner or admin can update
	if role != models.RoleAdmin && project.OwnerID != userID {
		return nil, errors.NewForbiddenError("You do not have permission to update this project")
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, errors.NewInternalServerError("Failed to update project", err)
	}

	s.invalidateCache(ctx)

	taskCount, _ := s.projectRepo.CountTasks(ctx, id)
	return toProjectResponse(project, taskCount), nil
}

func (s *projectService) Delete(ctx context.Context, id uint, userID uint, role string) *errors.AppError {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("Project not found")
		}
		return errors.NewInternalServerError("Failed to find project", err)
	}

	// Authorization check: Only project owner or admin can delete
	if role != models.RoleAdmin && project.OwnerID != userID {
		return errors.NewForbiddenError("You do not have permission to delete this project")
	}

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return errors.NewInternalServerError("Failed to delete project", err)
	}

	s.invalidateCache(ctx)
	return nil
}

func (s *projectService) List(ctx context.Context, pagination utils.PaginationQuery, ownerID *uint) ([]dto.ProjectResponse, int64, *errors.AppError) {
	projects, total, err := s.projectRepo.List(ctx, pagination, ownerID)
	if err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve projects", err)
	}

	var responses []dto.ProjectResponse
	for _, p := range projects {
		taskCount, _ := s.projectRepo.CountTasks(ctx, p.ID)
		responses = append(responses, *toProjectResponse(&p, taskCount))
	}

	return responses, total, nil
}

func (s *projectService) invalidateCache(ctx context.Context) {
	if s.cache != nil {
		_ = s.cache.DeleteByPattern(ctx, "project:*")
	}
}

func toProjectResponse(p *models.Project, taskCount int) *dto.ProjectResponse {
	if p == nil {
		return nil
	}
	return &dto.ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		OwnerID:     p.OwnerID,
		Owner:       toUserResponse(p.Owner),
		TasksCount:  taskCount,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
