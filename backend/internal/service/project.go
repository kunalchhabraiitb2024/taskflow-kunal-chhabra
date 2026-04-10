package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/model"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/realtime"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/repository"
)

// ErrForbidden is returned when an authenticated user lacks permission.
var ErrForbidden = errors.New("forbidden")

type ProjectService struct {
	projects *repository.ProjectRepository
	tasks    *repository.TaskRepository
	broker   *realtime.Broker
}

func NewProjectService(projects *repository.ProjectRepository, tasks *repository.TaskRepository, broker *realtime.Broker) *ProjectService {
	return &ProjectService{projects: projects, tasks: tasks, broker: broker}
}

func (s *ProjectService) List(ctx context.Context, userID uuid.UUID, p model.PaginationParams) (model.PaginatedResult[model.Project], error) {
	total, err := s.projects.CountByUser(ctx, userID)
	if err != nil {
		return model.PaginatedResult[model.Project]{}, fmt.Errorf("count projects: %w", err)
	}
	projects, err := s.projects.ListByUser(ctx, userID, p)
	if err != nil {
		return model.PaginatedResult[model.Project]{}, err
	}
	return model.NewPaginatedResult(projects, total, p), nil
}

func (s *ProjectService) Create(ctx context.Context, userID uuid.UUID, name string, description *string) (*model.Project, error) {
	return s.projects.Create(ctx, name, description, userID)
}

// GetByIDWithTasks returns the project and its tasks (for the detail view).
func (s *ProjectService) GetByIDWithTasks(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	project, err := s.projects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tasks, err := s.tasks.ListByProject(ctx, id, nil, nil, model.PaginationParams{Page: 1, Limit: 10000})
	if err != nil {
		return nil, fmt.Errorf("list tasks for project: %w", err)
	}
	if tasks == nil {
		tasks = []model.Task{}
	}
	project.Tasks = tasks
	return project, nil
}

func (s *ProjectService) Update(ctx context.Context, id, callerID uuid.UUID, name, description *string) (*model.Project, error) {
	project, err := s.projects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project.OwnerID != callerID {
		return nil, ErrForbidden
	}
	return s.projects.Update(ctx, id, name, description)
}

func (s *ProjectService) Delete(ctx context.Context, id, callerID uuid.UUID) error {
	project, err := s.projects.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if project.OwnerID != callerID {
		return ErrForbidden
	}
	if err := s.projects.Delete(ctx, id); err != nil {
		return err
	}
	if s.broker != nil {
		s.broker.PublishTasksChanged(id)
	}
	return nil
}
