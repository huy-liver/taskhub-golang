package dto

import "time"

// CreateTaskRequest represents request body for creating a task
type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,min=3,max=200"`
	Description string     `json:"description" binding:"omitempty,max=2000"`
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"due_date" binding:"omitempty"`
	ProjectID   uint       `json:"project_id" binding:"required"`
	AssigneeID  *uint      `json:"assignee_id" binding:"omitempty"`
}

// UpdateTaskRequest represents request body for updating a task
type UpdateTaskRequest struct {
	Title       string     `json:"title" binding:"omitempty,min=3,max=200"`
	Description string     `json:"description" binding:"omitempty,max=2000"`
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"due_date" binding:"omitempty"`
	AssigneeID  *uint      `json:"assignee_id" binding:"omitempty"`
}

// PatchTaskStatusRequest represents request body for updating task status
type PatchTaskStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=todo in_progress done"`
}

// TaskResponse represents task data returned to client
type TaskResponse struct {
	ID          uint             `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      string           `json:"status"`
	Priority    string           `json:"priority"`
	DueDate     *time.Time       `json:"due_date"`
	ProjectID   uint             `json:"project_id"`
	Project     *ProjectResponse `json:"project,omitempty"`
	AssigneeID  *uint            `json:"assignee_id"`
	Assignee    *UserResponse    `json:"assignee,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// TaskFilterQuery represents filter parameters for tasks
type TaskFilterQuery struct {
	ProjectID  *uint  `form:"project_id"`
	AssigneeID *uint  `form:"assignee_id"`
	Status     string `form:"status"`
	Priority   string `form:"priority"`
}
