package handlers

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"

	"taskflow/internal/middleware"
	"taskflow/internal/models"
	"taskflow/internal/repository"
	"taskflow/internal/validator"
)

// TaskHandler handles task endpoints.
type TaskHandler struct {
	taskRepo    *repository.TaskRepository
	projectRepo *repository.ProjectRepository
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(taskRepo *repository.TaskRepository, projectRepo *repository.ProjectRepository) *TaskHandler {
	return &TaskHandler{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
	}
}

// List handles GET /api/projects/:id/tasks — list tasks with filters and pagination.
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Verify project exists
	project, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil {
		slog.Error("failed to get project", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if project == nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}

	// Parse filters
	var status, assignee *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	if a := r.URL.Query().Get("assignee"); a != "" {
		assignee = &a
	}

	page, limit := parsePagination(r)

	tasks, totalCount, err := h.taskRepo.ListByProject(r.Context(), projectID, status, assignee, page, limit)
	if err != nil {
		slog.Error("failed to list tasks", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	respondJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       tasks,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(limit))),
	})
}

// Create handles POST /api/projects/:id/tasks — create a task in a project.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	// Verify project exists
	project, err := h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil {
		slog.Error("failed to get project", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if project == nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if ve := validator.ValidateCreateTask(&req); ve != nil {
		respondValidationError(w, ve.Fields)
		return
	}

	// Set defaults for optional fields
	status := models.TaskStatusTodo
	if req.Status != nil {
		status = *req.Status
	}
	priority := models.TaskPriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}

	task, err := h.taskRepo.Create(
		r.Context(),
		req.Title, req.Description,
		status, priority,
		projectID, req.AssigneeID,
		userID, req.DueDate,
	)
	if err != nil {
		slog.Error("failed to create task", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("task created", "task_id", task.ID, "project_id", projectID, "created_by", userID)
	respondJSON(w, http.StatusCreated, task)
}

// Update handles PATCH /api/tasks/:id — update a task.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	// Verify task exists
	task, err := h.taskRepo.GetByID(r.Context(), taskID)
	if err != nil {
		slog.Error("failed to get task", "error", err, "task_id", taskID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if task == nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}

	userID := middleware.GetUserID(r.Context())

	// Check authorization: project owner OR task assignee
	project, err := h.projectRepo.GetByID(r.Context(), task.ProjectID)
	if err != nil {
		slog.Error("failed to get project", "error", err, "project_id", task.ProjectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	isProjectOwner := project != nil && project.OwnerID == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID

	if !isProjectOwner && !isAssignee {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Parse body as map for PATCH semantics
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates, ve := validator.ValidateUpdateTask(body)
	if ve != nil {
		respondValidationError(w, ve.Fields)
		return
	}

	// Restrict modifications for assignees who are not the project owner
	if isAssignee && !isProjectOwner {
		restrictedUpdates := make(map[string]interface{})
		if status, ok := updates["status"]; ok {
			restrictedUpdates["status"] = status
		}
		updates = restrictedUpdates
	}

	if len(updates) == 0 {
		respondJSON(w, http.StatusOK, task)
		return
	}

	updated, err := h.taskRepo.Update(r.Context(), taskID, updates)
	if err != nil {
		slog.Error("failed to update task", "error", err, "task_id", taskID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("task updated", "task_id", taskID)
	respondJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/tasks/:id — delete a task (project owner or task creator only).
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	// Verify task exists
	task, err := h.taskRepo.GetByID(r.Context(), taskID)
	if err != nil {
		slog.Error("failed to get task", "error", err, "task_id", taskID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if task == nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}

	// Check authorization: project owner OR task creator
	project, err := h.projectRepo.GetByID(r.Context(), task.ProjectID)
	if err != nil {
		slog.Error("failed to get project", "error", err, "project_id", task.ProjectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	isProjectOwner := project != nil && project.OwnerID == userID
	isTaskCreator := task.CreatedBy == userID

	if !isProjectOwner && !isTaskCreator {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	deleted, err := h.taskRepo.Delete(r.Context(), taskID)
	if err != nil {
		slog.Error("failed to delete task", "error", err, "task_id", taskID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !deleted {
		respondError(w, http.StatusNotFound, "not found")
		return
	}

	slog.Info("task deleted", "task_id", taskID, "user_id", userID)
	w.WriteHeader(http.StatusNoContent)
}

// MyTasks handles GET /api/tasks/me — returns tasks assigned to the current user.
func (h *TaskHandler) MyTasks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	tasks, err := h.taskRepo.ListByAssignee(r.Context(), userID)
	if err != nil {
		slog.Error("failed to list assigned tasks", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	respondJSON(w, http.StatusOK, tasks)
}
