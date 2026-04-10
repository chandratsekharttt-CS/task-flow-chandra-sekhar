package models

import "time"

// Project represents a project owned by a user.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProjectWithTasks includes the project's tasks in the response.
type ProjectWithTasks struct {
	Project
	Tasks []Task `json:"tasks"`
}

// CreateProjectRequest is the payload for POST /projects.
type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// UpdateProjectRequest is the payload for PATCH /projects/:id.
type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// ProjectStats holds aggregated task statistics for a project.
type ProjectStats struct {
	ByStatus   map[string]int   `json:"by_status"`
	ByAssignee []AssigneeCount  `json:"by_assignee"`
	Total      int              `json:"total"`
}

// AssigneeCount holds a count of tasks assigned to a specific user.
type AssigneeCount struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Count  int    `json:"count"`
}
