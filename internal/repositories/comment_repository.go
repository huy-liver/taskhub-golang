package repositories

import (
	"context"
	"taskhub/internal/models"
	"taskhub/internal/utils"

	"gorm.io/gorm"
)

// CommentRepository defines interface for comment data operations
type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	GetByID(ctx context.Context, id uint) (*models.Comment, error)
	Update(ctx context.Context, comment *models.Comment) error
	Delete(ctx context.Context, id uint) error
	ListByTask(ctx context.Context, taskID uint, pagination utils.PaginationQuery) ([]models.Comment, int64, error)
}

type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository creates a new CommentRepository instance
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) GetByID(ctx context.Context, id uint) (*models.Comment, error) {
	var comment models.Comment
	if err := r.db.WithContext(ctx).Preload("User").First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

func (r *commentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Comment{}, id).Error
}

func (r *commentRepository) ListByTask(ctx context.Context, taskID uint, pagination utils.PaginationQuery) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("task_id = ?", taskID).
		Preload("User")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := pagination.SortBy + " " + pagination.SortDir
	err := query.Order(orderClause).
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Find(&comments).Error

	return comments, total, err
}
