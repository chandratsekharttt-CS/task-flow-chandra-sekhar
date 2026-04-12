package models

import "time"

// TaskStatus represents the current state of a task.
type TaskStatus string

// Valid task status values.
const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

// TaskPriority represents the importance of a task.
type TaskPriority string

// Valid task priority values.
const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// ValidStatuses is the set of allowed task status values.
var ValidStatuses = map[TaskStatus]bool{
	TaskStatusTodo:       true,
	TaskStatusInProgress: true,
	TaskStatusDone:       true,
}

// ValidPriorities is the set of allowed task priority values.
var ValidPriorities = map[TaskPriority]bool{
	TaskPriorityLow:    true,
	TaskPriorityMedium: true,
	TaskPriorityHigh:   true,
}

// Task represents a task within a project.
type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	ProjectID   string     `json:"project_id"`
	AssigneeID  *string    `json:"assignee_id"`
	CreatedBy        string     `json:"created_by"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	AssigneeName     *string    `json:"assignee_name,omitempty"`
	ProjectName      *string    `json:"project_name,omitempty"`
	ProjectOwnerName *string    `json:"project_owner_name,omitempty"`
}

// CreateTaskRequest is the payload for POST /projects/:id/tasks.
type CreateTaskRequest struct {
	Title       string        `json:"title"`
	Description *string       `json:"description"`
	Status      *TaskStatus   `json:"status"`
	Priority    *TaskPriority `json:"priority"`
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
