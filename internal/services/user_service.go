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

// UserService defines interface for user business operations
type UserService interface {
	Create(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, *errors.AppError)
	GetByID(ctx context.Context, id uint) (*dto.UserResponse, *errors.AppError)
	Update(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, *errors.AppError)
	Delete(ctx context.Context, id uint) *errors.AppError
	List(ctx context.Context, pagination utils.PaginationQuery) ([]dto.UserResponse, int64, *errors.AppError)
}

type userService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Create(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, *errors.AppError) {
	// Check email uniqueness
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, errors.NewConflictError("Email is already registered")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to hash password", err)
	}

	role := models.RoleUser
	if req.Role != "" {
		role = req.Role
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.NewInternalServerError("Failed to create user", err)
	}

	return toUserResponse(user), nil
}

func (s *userService) GetByID(ctx context.Context, id uint) (*dto.UserResponse, *errors.AppError) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("User not found")
		}
		return nil, errors.NewInternalServerError("Failed to get user", err)
	}
	return toUserResponse(user), nil
}

func (s *userService) Update(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, *errors.AppError) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("User not found")
		}
		return nil, errors.NewInternalServerError("Failed to get user", err)
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.NewInternalServerError("Failed to update user", err)
	}

	return toUserResponse(user), nil
}

func (s *userService) Delete(ctx context.Context, id uint) *errors.AppError {
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("User not found")
		}
		return errors.NewInternalServerError("Failed to find user", err)
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return errors.NewInternalServerError("Failed to delete user", err)
	}
	return nil
}

func (s *userService) List(ctx context.Context, pagination utils.PaginationQuery) ([]dto.UserResponse, int64, *errors.AppError) {
	users, total, err := s.userRepo.List(ctx, pagination)
	if err != nil {
		return nil, 0, errors.NewInternalServerError("Failed to retrieve users", err)
	}

	var responses []dto.UserResponse
	for _, u := range users {
		responses = append(responses, *toUserResponse(&u))
	}

	return responses, total, nil
}

func toUserResponse(u *models.User) *dto.UserResponse {
	if u == nil {
		return nil
	}
	return &dto.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
