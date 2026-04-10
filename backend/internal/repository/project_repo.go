package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/internal/models"
)

// ProjectRepository handles all project-related database operations.
type ProjectRepository struct {
	pool *pgxpool.Pool
}

// NewProjectRepository creates a new ProjectRepository.
func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

// Create inserts a new project.
func (r *ProjectRepository) Create(ctx context.Context, name string, description *string, ownerID string) (*models.Project, error) {
	var project models.Project
	err := r.pool.QueryRow(ctx,
		`INSERT INTO projects (name, description, owner_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, description, owner_id, created_at`,
		name, description, ownerID,
	).Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ListByUser returns projects the user owns or has tasks assigned in, with pagination.
func (r *ProjectRepository) ListByUser(ctx context.Context, userID string, page, limit int) ([]models.Project, int, error) {
	offset := (page - 1) * limit

	// Count total
	var totalCount int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT p.id)
		 FROM projects p
		 LEFT JOIN tasks t ON t.project_id = p.id
		 WHERE p.owner_id = $1 OR t.assignee_id = $1`,
		userID,
	).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		 FROM projects p
		 LEFT JOIN tasks t ON t.project_id = p.id
		 WHERE p.owner_id = $1 OR t.assignee_id = $1
		 ORDER BY p.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return projects, totalCount, nil
}

// GetByID retrieves a project by ID. Returns nil if not found.
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*models.Project, error) {
	var project models.Project
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, owner_id, created_at
		 FROM projects WHERE id = $1`,
		id,
	).Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

// Update applies partial updates to a project using a dynamic query.
func (r *ProjectRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*models.Project, error) {
	if len(updates) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
		args = append(args, value)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE projects SET %s WHERE id = $%d
		 RETURNING id, name, description, owner_id, created_at`,
		strings.Join(setClauses, ", "),
		argIdx,
	)

	var project models.Project
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

// Delete removes a project. Returns true if a row was deleted.
func (r *ProjectRepository) Delete(ctx context.Context, id string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// GetStats returns task count aggregations for a project.
func (r *ProjectRepository) GetStats(ctx context.Context, projectID string) (*models.ProjectStats, error) {
	stats := &models.ProjectStats{
		ByStatus: make(map[string]int),
	}

	// Count by status
	rows, err := r.pool.Query(ctx,
		`SELECT status::text, COUNT(*) FROM tasks
		 WHERE project_id = $1 GROUP BY status`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.ByStatus[status] = count
		stats.Total += count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Ensure all statuses are present
	for _, s := range []string{"todo", "in_progress", "done"} {
		if _, ok := stats.ByStatus[s]; !ok {
			stats.ByStatus[s] = 0
		}
	}

	// Count by assignee
	assigneeRows, err := r.pool.Query(ctx,
		`SELECT u.id, u.name, COUNT(t.id)
		 FROM tasks t
		 JOIN users u ON u.id = t.assignee_id
		 WHERE t.project_id = $1
		 GROUP BY u.id, u.name
		 ORDER BY COUNT(t.id) DESC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer assigneeRows.Close()

	for assigneeRows.Next() {
		var ac models.AssigneeCount
		if err := assigneeRows.Scan(&ac.UserID, &ac.Name, &ac.Count); err != nil {
			return nil, err
		}
		stats.ByAssignee = append(stats.ByAssignee, ac)
	}
	if err := assigneeRows.Err(); err != nil {
		return nil, err
	}

	if stats.ByAssignee == nil {
		stats.ByAssignee = []models.AssigneeCount{}
	}

	return stats, nil
}
