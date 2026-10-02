package routes

import (
	"taskhub/internal/configs"
	"taskhub/internal/handlers"
	"taskhub/internal/middlewares"
	"taskhub/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handlers holds all handlers to register routes
type Handlers struct {
	AuthHandler    *handlers.AuthHandler
	UserHandler    *handlers.UserHandler
	ProjectHandler *handlers.ProjectHandler
	TaskHandler    *handlers.TaskHandler
	CommentHandler *handlers.CommentHandler
}

// SetupRouter sets up Gin engine and registers all middleware and routes
func SetupRouter(cfg *configs.Config, logger *zap.Logger, h *Handlers) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// Global middlewares
	r.Use(middlewares.RequestID())
	r.Use(middlewares.ZapLogger(logger))
	r.Use(middlewares.RecoveryMiddleware(logger))
	r.Use(middlewares.CORSMiddleware(cfg.CORSAllowedOrigins))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"app":     cfg.AppName,
			"version": "1.0.0",
		})
	})

	apiV1 := r.Group("/api/v1")
	{
		// Public Auth routes
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/register", h.AuthHandler.Register)
			authGroup.POST("/login", h.AuthHandler.Login)
			authGroup.POST("/refresh", h.AuthHandler.RefreshToken)
		}

		// Protected routes
		protected := apiV1.Group("")
		protected.Use(middlewares.AuthMiddleware(cfg.JWTSecret))
		{
			// Auth me
			protected.GET("/auth/me", h.AuthHandler.GetMe)

			// Users routes
			usersGroup := protected.Group("/users")
			{
				usersGroup.GET("", h.UserHandler.List)
				usersGroup.GET("/:id", h.UserHandler.GetByID)
				usersGroup.PUT("/:id", h.UserHandler.Update)
				// Admin-only user management
				usersGroup.POST("", middlewares.RequireRole(models.RoleAdmin), h.UserHandler.Create)
				usersGroup.DELETE("/:id", middlewares.RequireRole(models.RoleAdmin), h.UserHandler.Delete)
			}

			// Projects routes
			projectsGroup := protected.Group("/projects")
			{
				projectsGroup.GET("", h.ProjectHandler.List)
				projectsGroup.POST("", h.ProjectHandler.Create)
				projectsGroup.GET("/:id", h.ProjectHandler.GetByID)
				projectsGroup.PUT("/:id", h.ProjectHandler.Update)
				projectsGroup.DELETE("/:id", h.ProjectHandler.Delete)
			}

			// Tasks routes
			tasksGroup := protected.Group("/tasks")
			{
				tasksGroup.GET("", h.TaskHandler.List)
				tasksGroup.POST("", h.TaskHandler.Create)
				tasksGroup.GET("/:id", h.TaskHandler.GetByID)
				tasksGroup.PUT("/:id", h.TaskHandler.Update)
				tasksGroup.PATCH("/:id/status", h.TaskHandler.PatchStatus)
				tasksGroup.DELETE("/:id", h.TaskHandler.Delete)

				// Comments under task
				tasksGroup.GET("/:id/comments", h.CommentHandler.ListByTask)
				tasksGroup.POST("/:id/comments", h.CommentHandler.Create)
			}

			// Comments routes
			commentsGroup := protected.Group("/comments")
			{
				commentsGroup.GET("", h.CommentHandler.ListByTask)
				commentsGroup.POST("", h.CommentHandler.Create)
				commentsGroup.GET("/:id", h.CommentHandler.GetByID)
				commentsGroup.PUT("/:id", h.CommentHandler.Update)
				commentsGroup.DELETE("/:id", h.CommentHandler.Delete)
			}
		}
	}

	return r
}
