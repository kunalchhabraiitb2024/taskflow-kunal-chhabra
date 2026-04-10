package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/model"
)

// listByUserBaseQuery is the common WHERE clause reused for both count and paginated select.
const listByUserBaseQuery = `
	FROM projects p
	LEFT JOIN tasks t ON t.project_id = p.id
	WHERE p.owner_id = $1 OR t.assignee_id = $1`

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

func (r *ProjectRepository) Create(ctx context.Context, name string, description *string, ownerID uuid.UUID) (*model.Project, error) {
	p := &model.Project{Name: name, Description: description, OwnerID: ownerID}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO projects (name, description, owner_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		name, description, ownerID,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	p := &model.Project{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

// CountByUser returns total number of distinct projects visible to this user.
func (r *ProjectRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT p.id)`+listByUserBaseQuery,
		userID,
	).Scan(&total)
	return total, err
}

// ListByUser returns paginated projects the user owns OR has tasks assigned to them in.
func (r *ProjectRepository) ListByUser(ctx context.Context, userID uuid.UUID, p model.PaginationParams) ([]model.Project, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at`+
			listByUserBaseQuery+
			` ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`,
		userID, p.Limit, p.Offset(),
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var proj model.Project
		if err := rows.Scan(&proj.ID, &proj.Name, &proj.Description, &proj.OwnerID, &proj.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, proj)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, name *string, description *string) (*model.Project, error) {
	p := &model.Project{}
	err := r.pool.QueryRow(ctx,
		`UPDATE projects
		 SET name        = COALESCE($2, name),
		     description = COALESCE($3, description)
		 WHERE id = $1
		 RETURNING id, name, description, owner_id, created_at`,
		id, name, description,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	return p, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
