package services

import (
	"context"
	"taskhub/internal/configs"
	"taskhub/internal/dto"
	"taskhub/internal/errors"
	"taskhub/internal/models"
	"taskhub/internal/repositories"
	"taskhub/internal/utils"

	"gorm.io/gorm"
)

// AuthService defines interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, *errors.AppError)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, *errors.AppError)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, *errors.AppError)
	GetMe(ctx context.Context, userID uint) (*dto.UserResponse, *errors.AppError)
}

type authService struct {
	userRepo repositories.UserRepository
	cfg      *configs.Config
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repositories.UserRepository, cfg *configs.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, *errors.AppError) {
	// Check if email already registered
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

	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(user.ID, user.Role, s.cfg)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to generate authentication tokens", err)
	}

	return &dto.AuthResponse{
		User:         *toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, *errors.AppError) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewUnauthorizedError("Invalid email or password")
		}
		return nil, errors.NewInternalServerError("Failed to verify user", err)
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.NewUnauthorizedError("Invalid email or password")
	}

	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(user.ID, user.Role, s.cfg)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to generate authentication tokens", err)
	}

	return &dto.AuthResponse{
		User:         *toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, *errors.AppError) {
	claims, err := utils.ValidateToken(req.RefreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, errors.NewUnauthorizedError("Invalid or expired refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewUnauthorizedError("User no longer exists")
		}
		return nil, errors.NewInternalServerError("Failed to verify user", err)
	}

	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(user.ID, user.Role, s.cfg)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to generate authentication tokens", err)
	}

	return &dto.AuthResponse{
		User:         *toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID uint) (*dto.UserResponse, *errors.AppError) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("User not found")
		}
		return nil, errors.NewInternalServerError("Failed to get current user", err)
	}

	return toUserResponse(user), nil
}
