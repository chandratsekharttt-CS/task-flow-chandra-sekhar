package validator

import (
	"net/mail"
	"strings"

	"taskflow/internal/models"
)

// ValidationError holds structured field-level validation errors.
type ValidationError struct {
	Fields map[string]string `json:"fields"`
}

func (v *ValidationError) Error() string {
	return "validation failed"
}

// HasErrors returns true if there are validation errors.
func (v *ValidationError) HasErrors() bool {
	return len(v.Fields) > 0
}

// NewValidationError creates a new empty ValidationError.
func NewValidationError() *ValidationError {
	return &ValidationError{Fields: make(map[string]string)}
}

// ValidateRegister validates a user registration request.
func ValidateRegister(req *models.RegisterRequest) *ValidationError {
	ve := NewValidationError()

	if strings.TrimSpace(req.Name) == "" {
		ve.Fields["name"] = "is required"
	}
	if strings.TrimSpace(req.Email) == "" {
		ve.Fields["email"] = "is required"
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		ve.Fields["email"] = "is not a valid email address"
	}
	if len(req.Password) < 8 {
		ve.Fields["password"] = "must be at least 8 characters"
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// ValidateLogin validates a login request.
func ValidateLogin(req *models.LoginRequest) *ValidationError {
	ve := NewValidationError()

	if strings.TrimSpace(req.Email) == "" {
		ve.Fields["email"] = "is required"
	}
	if strings.TrimSpace(req.Password) == "" {
		ve.Fields["password"] = "is required"
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// ValidateCreateProject validates a project creation request.
func ValidateCreateProject(req *models.CreateProjectRequest) *ValidationError {
	ve := NewValidationError()

	if strings.TrimSpace(req.Name) == "" {
		ve.Fields["name"] = "is required"
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// ValidateCreateTask validates a task creation request.
func ValidateCreateTask(req *models.CreateTaskRequest) *ValidationError {
	ve := NewValidationError()

	if strings.TrimSpace(req.Title) == "" {
		ve.Fields["title"] = "is required"
	}
	if req.Status != nil && !models.ValidStatuses[*req.Status] {
		ve.Fields["status"] = "must be one of: todo, in_progress, done"
	}
	if req.Priority != nil && !models.ValidPriorities[*req.Priority] {
		ve.Fields["priority"] = "must be one of: low, medium, high"
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// ValidateUpdateTask validates fields in a task update request (map-based for PATCH semantics).
func ValidateUpdateTask(body map[string]interface{}) (map[string]interface{}, *ValidationError) {
	ve := NewValidationError()
	updates := make(map[string]interface{})

	if title, ok := body["title"]; ok {
		s, isStr := title.(string)
		if !isStr || strings.TrimSpace(s) == "" {
			ve.Fields["title"] = "must be a non-empty string"
		} else {
			updates["title"] = s
		}
	}
	if desc, ok := body["description"]; ok {
		if desc == nil {
			updates["description"] = nil
		} else if s, isStr := desc.(string); isStr {
			updates["description"] = s
		} else {
			ve.Fields["description"] = "must be a string or null"
		}
	}
	if status, ok := body["status"]; ok {
		s, isStr := status.(string)
		if !isStr || !models.ValidStatuses[s] {
			ve.Fields["status"] = "must be one of: todo, in_progress, done"
		} else {
			updates["status"] = s
		}
	}
	if priority, ok := body["priority"]; ok {
		s, isStr := priority.(string)
		if !isStr || !models.ValidPriorities[s] {
			ve.Fields["priority"] = "must be one of: low, medium, high"
		} else {
			updates["priority"] = s
		}
	}
	if assignee, ok := body["assignee_id"]; ok {
		if assignee == nil {
			updates["assignee_id"] = nil
		} else if s, isStr := assignee.(string); isStr && s != "" {
			updates["assignee_id"] = s
		} else if s, isStr := assignee.(string); isStr && s == "" {
			updates["assignee_id"] = nil
		} else {
			ve.Fields["assignee_id"] = "must be a valid user UUID or null"
		}
	}
	if dueDate, ok := body["due_date"]; ok {
		if dueDate == nil {
			updates["due_date"] = nil
		} else if s, isStr := dueDate.(string); isStr {
			updates["due_date"] = s
		} else {
			ve.Fields["due_date"] = "must be a date string (YYYY-MM-DD) or null"
		}
	}

	if ve.HasErrors() {
		return nil, ve
	}
	return updates, nil
}

// ValidateUpdateProject validates fields in a project update request.
func ValidateUpdateProject(body map[string]interface{}) (map[string]interface{}, *ValidationError) {
	ve := NewValidationError()
	updates := make(map[string]interface{})

	if name, ok := body["name"]; ok {
		s, isStr := name.(string)
		if !isStr || strings.TrimSpace(s) == "" {
			ve.Fields["name"] = "must be a non-empty string"
		} else {
			updates["name"] = s
		}
	}
	if desc, ok := body["description"]; ok {
		if desc == nil {
			updates["description"] = nil
		} else if s, isStr := desc.(string); isStr {
			updates["description"] = s
		} else {
			ve.Fields["description"] = "must be a string or null"
		}
	}

	if ve.HasErrors() {
		return nil, ve
	}
	return updates, nil
}
