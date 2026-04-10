package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/model"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, t *model.Task) (*model.Task, error) {
	out := &model.Task{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, created_by, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at`,
		t.Title, t.Description, t.Status, t.Priority, t.ProjectID, t.AssigneeID, t.CreatedBy, t.DueDate,
	).Scan(
		&out.ID, &out.Title, &out.Description, &out.Status, &out.Priority,
		&out.ProjectID, &out.AssigneeID, &out.CreatedBy, &out.DueDate,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return out, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	t := &model.Task{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at
		 FROM tasks WHERE id = $1`,
		id,
	).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

const taskFilterBase = `
	FROM tasks
	WHERE project_id = $1
	  AND ($2::task_status IS NULL OR status = $2)
	  AND ($3::uuid IS NULL OR assignee_id = $3)`

const taskSelectCols = `SELECT id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at`

// CountByProject returns the total number of tasks matching the given filters.
func (r *TaskRepository) CountByProject(ctx context.Context, projectID uuid.UUID, status *model.TaskStatus, assigneeID *uuid.UUID) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)`+taskFilterBase,
		projectID, status, assigneeID,
	).Scan(&total)
	return total, err
}

// ListByProject lists tasks for a project with optional filters and pagination.
func (r *TaskRepository) ListByProject(ctx context.Context, projectID uuid.UUID, status *model.TaskStatus, assigneeID *uuid.UUID, p model.PaginationParams) ([]model.Task, error) {
	rows, err := r.pool.Query(ctx,
		taskSelectCols+taskFilterBase+` ORDER BY created_at DESC LIMIT $4 OFFSET $5`,
		projectID, status, assigneeID, p.Limit, p.Offset(),
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// TaskUpdate holds the fields that can be updated on a task.
// Nil pointer = "don't change this field".
// ClearAssigneeID / ClearDueDate = explicitly set the field to NULL.
type TaskUpdate struct {
	Title          *string
	Description    *string
	Status         *model.TaskStatus
	Priority       *model.TaskPriority
	AssigneeID     *uuid.UUID
	ClearAssigneeID bool // if true, set assignee_id = NULL
	DueDate        *time.Time
	ClearDueDate   bool // if true, set due_date = NULL
}

func (r *TaskRepository) Update(ctx context.Context, id uuid.UUID, u TaskUpdate) (*model.Task, error) {
	t := &model.Task{}
	err := r.pool.QueryRow(ctx,
		`UPDATE tasks SET
		   title       = COALESCE($2, title),
		   description = COALESCE($3, description),
		   status      = COALESCE($4, status),
		   priority    = COALESCE($5, priority),
		   assignee_id = CASE WHEN $6 THEN NULL ELSE COALESCE($7, assignee_id) END,
		   due_date    = CASE WHEN $8 THEN NULL ELSE COALESCE($9, due_date) END,
		   updated_at  = NOW()
		 WHERE id = $1
		 RETURNING id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at`,
		id,
		u.Title, u.Description, u.Status, u.Priority,
		u.ClearAssigneeID, u.AssigneeID,
		u.ClearDueDate, u.DueDate,
	).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
