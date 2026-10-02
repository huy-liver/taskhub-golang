package models

import (
	"time"

	"gorm.io/gorm"
)

// Role constants
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Task Status constants
const (
	TaskStatusTodo       = "todo"
	TaskStatusInProgress = "in_progress"
	TaskStatusDone       = "done"
)

// Task Priority constants
const (
	TaskPriorityLow    = "low"
	TaskPriorityMedium = "medium"
	TaskPriorityHigh   = "high"
)

// User represents a user in the system
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Role      string         `gorm:"size:20;not null;default:'user'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Projects []Project `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"projects,omitempty"`
	Tasks    []Task    `gorm:"foreignKey:AssigneeID;constraint:OnDelete:SET NULL" json:"tasks,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
}

// Project represents a project owned by a user
type Project struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:150;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	OwnerID     uint           `gorm:"not null;index" json:"owner_id"`
	Owner       *User          `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Tasks       []Task         `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"tasks,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Task represents an individual task within a project
type Task struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"size:30;not null;default:'todo';index" json:"status"`
	Priority    string         `gorm:"size:30;not null;default:'medium';index" json:"priority"`
	DueDate     *time.Time     `json:"due_date"`
	ProjectID   uint           `gorm:"not null;index" json:"project_id"`
	Project     *Project       `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	AssigneeID  *uint          `gorm:"index" json:"assignee_id"`
	Assignee    *User          `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Comments    []Comment      `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Comment represents a user comment on a task
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	TaskID    uint           `gorm:"not null;index" json:"task_id"`
	Task      *Task          `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
