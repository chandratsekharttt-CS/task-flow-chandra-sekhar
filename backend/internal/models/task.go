package models

import "time"

// Valid task status values.
const (
	TaskStatusTodo       = "todo"
	TaskStatusInProgress = "in_progress"
	TaskStatusDone       = "done"
)

// Valid task priority values.
const (
	TaskPriorityLow    = "low"
	TaskPriorityMedium = "medium"
	TaskPriorityHigh   = "high"
)

// ValidStatuses is the set of allowed task status values.
var ValidStatuses = map[string]bool{
	TaskStatusTodo:       true,
	TaskStatusInProgress: true,
	TaskStatusDone:       true,
}

// ValidPriorities is the set of allowed task priority values.
var ValidPriorities = map[string]bool{
	TaskPriorityLow:    true,
	TaskPriorityMedium: true,
	TaskPriorityHigh:   true,
}

// Task represents a task within a project.
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	ProjectID   string     `json:"project_id"`
	AssigneeID  *string    `json:"assignee_id"`
	CreatedBy   string     `json:"created_by"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateTaskRequest is the payload for POST /projects/:id/tasks.
type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"` // format: "2006-01-02"
}

// PaginationParams holds pagination query parameters.
type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// PaginatedResponse wraps a list response with pagination metadata.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}
