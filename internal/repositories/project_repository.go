package repositories

import (
	"context"
	"taskhub/internal/models"
	"taskhub/internal/utils"

	"gorm.io/gorm"
)

// ProjectRepository defines interface for project data operations
type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) error
	GetByID(ctx context.Context, id uint) (*models.Project, error)
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, pagination utils.PaginationQuery, ownerID *uint) ([]models.Project, int64, error)
	CountTasks(ctx context.Context, projectID uint) (int, error)
}

type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new ProjectRepository instance
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *projectRepository) GetByID(ctx context.Context, id uint) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).Preload("Owner").First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) Update(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *projectRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Project{}, id).Error
}

func (r *projectRepository) List(ctx context.Context, pagination utils.PaginationQuery, ownerID *uint) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Project{}).Preload("Owner")

	if ownerID != nil {
		query = query.Where("owner_id = ?", *ownerID)
	}

	if pagination.Search != "" {
		searchTerm := "%" + pagination.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := pagination.SortBy + " " + pagination.SortDir
	err := query.Order(orderClause).
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Find(&projects).Error

	return projects, total, err
}

func (r *projectRepository) CountTasks(ctx context.Context, projectID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Task{}).Where("project_id = ?", projectID).Count(&count).Error
	return int(count), err
}
