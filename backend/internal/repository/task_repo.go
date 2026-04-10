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

// TaskRepository handles all task-related database operations.
type TaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository creates a new TaskRepository.
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

// taskColumns is the standard column list for scanning tasks.
const taskColumns = `id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at`

// scanTask scans a row into a Task model.
func scanTask(row pgx.Row) (*models.Task, error) {
	var t models.Task
	err := row.Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Create inserts a new task.
func (r *TaskRepository) Create(ctx context.Context, title string, description *string, status, priority, projectID string, assigneeID *string, createdBy string, dueDate *string) (*models.Task, error) {
	query := fmt.Sprintf(
		`INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, created_by, due_date)
		 VALUES ($1, $2, $3::task_status, $4::task_priority, $5, $6, $7, $8::date)
		 RETURNING %s`, taskColumns,
	)
	row := r.pool.QueryRow(ctx, query,
		title, description, status, priority, projectID, assigneeID, createdBy, dueDate,
	)
	return scanTask(row)
}

// ListByProject returns tasks for a project with optional filters and pagination.
func (r *TaskRepository) ListByProject(ctx context.Context, projectID string, status, assignee *string, page, limit int) ([]models.Task, int, error) {
	offset := (page - 1) * limit

	// Build WHERE clause dynamically
	whereClauses := []string{"project_id = $1"}
	args := []interface{}{projectID}
	argIdx := 2

	if status != nil && *status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d::task_status", argIdx))
		args = append(args, *status)
		argIdx++
	}
	if assignee != nil && *assignee != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("assignee_id = $%d", argIdx))
		args = append(args, *assignee)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Count total matching
	var totalCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", whereSQL)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	listArgs := append(args, limit, offset)
	listQuery := fmt.Sprintf(
		"SELECT %s FROM tasks WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		taskColumns, whereSQL, argIdx, argIdx+1,
	)
	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, totalCount, nil
}

// GetByID retrieves a task by ID. Returns nil if not found.
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	query := fmt.Sprintf("SELECT %s FROM tasks WHERE id = $1", taskColumns)
	task, err := scanTask(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return task, nil
}

// Update applies partial updates to a task using a dynamic query.
func (r *TaskRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*models.Task, error) {
	if len(updates) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	// Map of field names to their SQL type casts
	typeCasts := map[string]string{
		"status":   "::task_status",
		"priority": "::task_priority",
		"due_date": "::date",
	}

	for field, value := range updates {
		cast := typeCasts[field]
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d%s", field, argIdx, cast))
		args = append(args, value)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE tasks SET %s WHERE id = $%d RETURNING %s",
		strings.Join(setClauses, ", "),
		argIdx,
		taskColumns,
	)

	task, err := scanTask(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return task, nil
}

// Delete removes a task. Returns true if a row was deleted.
func (r *TaskRepository) Delete(ctx context.Context, id string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// ListByProjectNoPage returns all tasks for a project (used for project detail).
func (r *TaskRepository) ListByProjectNoPage(ctx context.Context, projectID string) ([]models.Task, error) {
	query := fmt.Sprintf("SELECT %s FROM tasks WHERE project_id = $1 ORDER BY created_at DESC", taskColumns)
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ListByAssignee returns all tasks assigned to a specific user across all projects.
func (r *TaskRepository) ListByAssignee(ctx context.Context, assigneeID string) ([]models.Task, error) {
	query := fmt.Sprintf("SELECT %s FROM tasks WHERE assignee_id = $1 ORDER BY created_at DESC", taskColumns)
	rows, err := r.pool.Query(ctx, query, assigneeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
