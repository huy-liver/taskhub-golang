package dto

import "time"

// CreateProjectRequest represents request body for creating a project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=150"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// UpdateProjectRequest represents request body for updating a project
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"omitempty,min=3,max=150"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// ProjectResponse represents project data returned to client
type ProjectResponse struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	OwnerID     uint          `json:"owner_id"`
	Owner       *UserResponse `json:"owner,omitempty"`
	TasksCount  int           `json:"tasks_count,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
