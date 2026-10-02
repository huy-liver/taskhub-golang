package repositories

import (
	"context"
	"taskhub/internal/dto"
	"taskhub/internal/models"
	"taskhub/internal/utils"

	"gorm.io/gorm"
)

// TaskRepository defines interface for task data operations
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id uint) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	PatchStatus(ctx context.Context, id uint, status string) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, pagination utils.PaginationQuery, filter dto.TaskFilterQuery) ([]models.Task, int64, error)
}

type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new TaskRepository instance
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetByID(ctx context.Context, id uint) (*models.Task, error) {
	var task models.Task
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Project.Owner").
		Preload("Assignee").
		First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) Update(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *taskRepository) PatchStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Task{}).Where("id = ?", id).Update("status", status).Error
}

func (r *taskRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Task{}, id).Error
}

func (r *taskRepository) List(ctx context.Context, pagination utils.PaginationQuery, filter dto.TaskFilterQuery) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Task{}).
		Preload("Project").
		Preload("Assignee")

	if filter.ProjectID != nil {
		query = query.Where("project_id = ?", *filter.ProjectID)
	}

	if filter.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *filter.AssigneeID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}

	if pagination.Search != "" {
		searchTerm := "%" + pagination.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := pagination.SortBy + " " + pagination.SortDir
	err := query.Order(orderClause).
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Find(&tasks).Error

	return tasks, total, err
}
