package dto

import "time"

// CreateCommentRequest represents request body for creating a comment
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
	TaskID  uint   `json:"task_id" binding:"required"`
}

// UpdateCommentRequest represents request body for updating a comment
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// CommentResponse represents comment data returned to client
type CommentResponse struct {
	ID        uint          `json:"id"`
	Content   string        `json:"content"`
	TaskID    uint          `json:"task_id"`
	UserID    uint          `json:"user_id"`
	User      *UserResponse `json:"user,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
