package handlers

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"taskflow/internal/middleware"
	"taskflow/internal/models"
	"taskflow/internal/repository"
	"taskflow/internal/validator"
)

// ProjectHandler handles project endpoints.
type ProjectHandler struct {
	projectRepo *repository.ProjectRepository
	taskRepo    *repository.TaskRepository
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(projectRepo *repository.ProjectRepository, taskRepo *repository.TaskRepository) *ProjectHandler {
	return &ProjectHandler{
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
	}
}

// List handles GET /api/projects — list projects for the current user.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page, limit := parsePagination(r)

	projects, totalCount, err := h.projectRepo.ListByUser(r.Context(), userID, page, limit)
	if err != nil {
		slog.Error("failed to list projects", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if projects == nil {
		projects = []models.Project{}
	}

	respondJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       projects,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(limit))),
	})
}

// Create handles POST /api/projects — create a new project.
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if ve := validator.ValidateCreateProject(&req); ve != nil {
		respondValidationError(w, ve.Fields)
		return
	}

	userID := middleware.GetUserID(r.Context())
	project, err := h.projectRepo.Create(r.Context(), req.Name, req.Description, userID)
	if err != nil {
		slog.Error("failed to create project", "error", err, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("project created", "project_id", project.ID, "owner_id", userID)
	respondJSON(w, http.StatusCreated, project)
}

// Get handles GET /api/projects/:id — get project details with its tasks.
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

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

	tasks, err := h.taskRepo.ListByProjectNoPage(r.Context(), projectID)
	if err != nil {
		slog.Error("failed to list tasks", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	respondJSON(w, http.StatusOK, models.ProjectWithTasks{
		Project: *project,
		Tasks:   tasks,
	})
}

// Update handles PATCH /api/projects/:id — update project (owner only).
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	// Check project exists and user is owner
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
	if project.OwnerID != userID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Parse body as map for PATCH semantics
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates, ve := validator.ValidateUpdateProject(body)
	if ve != nil {
		respondValidationError(w, ve.Fields)
		return
	}
	if len(updates) == 0 {
		respondJSON(w, http.StatusOK, project)
		return
	}

	updated, err := h.projectRepo.Update(r.Context(), projectID, updates)
	if err != nil {
		slog.Error("failed to update project", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("project updated", "project_id", projectID, "user_id", userID)
	respondJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/projects/:id — delete project (owner only).
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	// Check project exists and user is owner
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
	if project.OwnerID != userID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	deleted, err := h.projectRepo.Delete(r.Context(), projectID)
	if err != nil {
		slog.Error("failed to delete project", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !deleted {
		respondError(w, http.StatusNotFound, "not found")
		return
	}

	slog.Info("project deleted", "project_id", projectID, "user_id", userID)
	w.WriteHeader(http.StatusNoContent)
}

// Stats handles GET /api/projects/:id/stats — task statistics for a project.
func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

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

	stats, err := h.projectRepo.GetStats(r.Context(), projectID)
	if err != nil {
		slog.Error("failed to get project stats", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// parsePagination extracts page and limit from query params with defaults.
func parsePagination(r *http.Request) (int, int) {
	page := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	return page, limit
}
