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

// CommentService defines interface for comment business operations
type CommentService interface {
	Create(ctx context.Context, userID uint, req *dto.CreateCommentRequest) (*dto.CommentResponse, *errors.AppError)
	GetByID(ctx context.Context, id uint) (*dto.CommentResponse, *errors.AppError)
	Update(ctx context.Context, id uint, userID uint, role string, req *dto.UpdateCommentRequest) (*dto.CommentResponse, *errors.AppError)
	Delete(ctx context.Context, id uint, userID uint, role string) *errors.AppError
	ListByTask(ctx context.Context, taskID uint, pagination utils.PaginationQuery) ([]dto.CommentResponse, int64, *errors.AppError)
}

type commentService struct {
	commentRepo repositories.CommentRepository
	taskRepo    repositories.TaskRepository
}

// NewCommentService creates a new CommentService instance
func NewCommentService(commentRepo repositories.CommentRepository, taskRepo repositories.TaskRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		taskRepo:    taskRepo,
	}
}

func (s *commentService) Create(ctx context.Context, userID uint, req *dto.CreateCommentRequest) (*dto.CommentResponse, *errors.AppError) {
	// Verify task exists
	_, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Task not found")
		}
		return nil, errors.NewInternalServerError("Failed to check task", err)
	}

	comment := &models.Comment{
		Content: req.Content,
		TaskID:  req.TaskID,
		UserID:  userID,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, errors.NewInternalServerError("Failed to create comment", err)
	}

	created, err := s.commentRepo.GetByID(ctx, comment.ID)
	if err != nil {
		return toCommentResponse(comment), nil
	}

	return toCommentResponse(created), nil
}

func (s *commentService) GetByID(ctx context.Context, id uint) (*dto.CommentResponse, *errors.AppError) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Comment not found")
		}
		return nil, errors.NewInternalServerError("Failed to get comment", err)
	}
	return toCommentResponse(comment), nil
}

func (s *commentService) Update(ctx context.Context, id uint, userID uint, role string, req *dto.UpdateCommentRequest) (*dto.CommentResponse, *errors.AppError) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Comment not found")
		}
		return nil, errors.NewInternalServerError("Failed to get comment", err)
	}

	// Authorization check: Only comment author or admin can update
	if role != models.RoleAdmin && comment.UserID != userID {
		return nil, errors.NewForbiddenError("You do not have permission to update this comment")
	}

	comment.Content = req.Content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, errors.NewInternalServerError("Failed to update comment", err)
	}

	return toCommentResponse(comment), nil
}

func (s *commentService) Delete(ctx context.Context, id uint, userID uint, role string) *errors.AppError {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("Comment not found")
		}
		return errors.NewInternalServerError("Failed to get comment", err)
	}

	// Authorization check: Only comment author or admin can delete
	if role != models.RoleAdmin && comment.UserID != userID {
		return errors.NewForbiddenError("You do not have permission to delete this comment")
	}

	if err := s.commentRepo.Delete(ctx, id); err != nil {
		return errors.NewInternalServerError("Failed to delete comment", err)
	}
	return nil
}

func (s *commentService) ListByTask(ctx context.Context, taskID uint, pagination utils.PaginationQuery) ([]dto.CommentResponse, int64, *errors.AppError) {
	comments, total, err := s.commentRepo.ListByTask(ctx, taskID, pagination)
	if err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve comments", err)
	}

	var responses []dto.CommentResponse
	for _, c := range comments {
		responses = append(responses, *toCommentResponse(&c))
	}
	return responses, total, nil
}

func toCommentResponse(c *models.Comment) *dto.CommentResponse {
	if c == nil {
		return nil
	}
	return &dto.CommentResponse{
		ID:        c.ID,
		Content:   c.Content,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		User:      toUserResponse(c.User),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
