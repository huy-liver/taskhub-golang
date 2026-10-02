package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskhub/internal/configs"
	"taskhub/internal/handlers"
	"taskhub/internal/models"
	"taskhub/internal/repositories"
	"taskhub/internal/routes"
	"taskhub/internal/services"
	"taskhub/internal/utils"

	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	cfg, err := configs.LoadConfig()
	if err != nil {
		fmt.Printf("Critical: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	logger, err := utils.InitLogger(cfg.AppEnv)
	if err != nil {
		fmt.Printf("Critical: failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("Starting TaskHub API Server...",
		zap.String("app_name", cfg.AppName),
		zap.String("environment", cfg.AppEnv),
		zap.String("port", cfg.AppPort),
	)

	// 3. Initialize Database
	db, err := models.InitDB(cfg)
	if err != nil {
		logger.Warn("Failed to connect to primary database (continuing in degraded mode for tests/docs)",
			zap.Error(err),
		)
	} else {
		logger.Info("Database connection and migrations completed successfully")
	}

	// 4. Initialize Repositories
	userRepo := repositories.NewUserRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	taskRepo := repositories.NewTaskRepository(db)
	commentRepo := repositories.NewCommentRepository(db)

	// 5. Initialize Services
	authService := services.NewAuthService(userRepo, cfg)
	userService := services.NewUserService(userRepo)
	projectService := services.NewProjectService(projectRepo)
	taskService := services.NewTaskService(taskRepo, projectRepo, userRepo)
	commentService := services.NewCommentService(commentRepo, taskRepo)

	// 6. Initialize Handlers
	h := &routes.Handlers{
		AuthHandler:    handlers.NewAuthHandler(authService),
		UserHandler:    handlers.NewUserHandler(userService),
		ProjectHandler: handlers.NewProjectHandler(projectService),
		TaskHandler:    handlers.NewTaskHandler(taskService),
		CommentHandler: handlers.NewCommentHandler(commentService),
	}

	// 7. Initialize Router
	router := routes.SetupRouter(cfg, logger, h)

	// 8. Start HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Server listening on port " + cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed to listen", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
