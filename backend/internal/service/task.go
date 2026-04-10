package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/model"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/realtime"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/repository"
)

type TaskService struct {
	tasks    *repository.TaskRepository
	projects *repository.ProjectRepository
	broker   *realtime.Broker
}

func NewTaskService(tasks *repository.TaskRepository, projects *repository.ProjectRepository, broker *realtime.Broker) *TaskService {
	return &TaskService{tasks: tasks, projects: projects, broker: broker}
}

func (s *TaskService) ListByProject(ctx context.Context, projectID uuid.UUID, status *model.TaskStatus, assigneeID *uuid.UUID, p model.PaginationParams) (model.PaginatedResult[model.Task], error) {
	total, err := s.tasks.CountByProject(ctx, projectID, status, assigneeID)
	if err != nil {
		return model.PaginatedResult[model.Task]{}, fmt.Errorf("count tasks: %w", err)
	}
	tasks, err := s.tasks.ListByProject(ctx, projectID, status, assigneeID, p)
	if err != nil {
		return model.PaginatedResult[model.Task]{}, err
	}
	return model.NewPaginatedResult(tasks, total, p), nil
}

func (s *TaskService) Create(ctx context.Context, callerID, projectID uuid.UUID, t *model.Task) (*model.Task, error) {
	// Verify project exists
	if _, err := s.projects.GetByID(ctx, projectID); err != nil {
		return nil, err
	}
	t.ProjectID = projectID
	t.CreatedBy = callerID
	if t.Status == "" {
		t.Status = model.StatusTodo
	}
	if t.Priority == "" {
		t.Priority = model.PriorityMedium
	}
	created, err := s.tasks.Create(ctx, t)
	if err != nil {
		return nil, err
	}
	if s.broker != nil {
		s.broker.PublishTasksChanged(projectID)
	}
	return created, nil
}

func (s *TaskService) Update(ctx context.Context, id uuid.UUID, u repository.TaskUpdate) (*model.Task, error) {
	if _, err := s.tasks.GetByID(ctx, id); err != nil {
		return nil, err
	}
	task, err := s.tasks.Update(ctx, id, u)
	if err != nil {
		return nil, err
	}
	if s.broker != nil {
		s.broker.PublishTasksChanged(task.ProjectID)
	}
	return task, nil
}

// Delete allows deletion by project owner OR task creator.
func (s *TaskService) Delete(ctx context.Context, id, callerID uuid.UUID) error {
	task, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if task.CreatedBy != callerID {
		project, err := s.projects.GetByID(ctx, task.ProjectID)
		if err != nil {
			return fmt.Errorf("get project for task: %w", err)
		}
		if project.OwnerID != callerID {
			return ErrForbidden
		}
	}

	if err := s.tasks.Delete(ctx, id); err != nil {
		return err
	}
	if s.broker != nil {
		s.broker.PublishTasksChanged(task.ProjectID)
	}
	return nil
}

// GetStats returns task count by status and by assignee for a project.
type TaskStats struct {
	ByStatus   map[string]int `json:"by_status"`
	ByAssignee []AssigneeStat `json:"by_assignee"`
	Total      int            `json:"total"`
}

type AssigneeStat struct {
	UserID *uuid.UUID `json:"user_id"`
	Name   string     `json:"name"`
	Count  int        `json:"count"`
}

func (s *TaskService) GetStats(ctx context.Context, projectID uuid.UUID) (*TaskStats, error) {
	// Stats needs all tasks — use a high limit to get everything
	allPages := model.PaginationParams{Page: 1, Limit: 10000}
	tasks, err := s.tasks.ListByProject(ctx, projectID, nil, nil, allPages)
	if err != nil {
		return nil, err
	}

	byStatus := map[string]int{"todo": 0, "in_progress": 0, "done": 0}
	byAssignee := map[string]*AssigneeStat{}

	for _, t := range tasks {
		byStatus[string(t.Status)]++

		key := "unassigned"
		if t.AssigneeID != nil {
			key = t.AssigneeID.String()
		}
		if _, ok := byAssignee[key]; !ok {
			stat := &AssigneeStat{UserID: t.AssigneeID, Name: "Unassigned"}
			if t.AssigneeID != nil {
				stat.Name = key // will be enriched later if needed
			}
			byAssignee[key] = stat
		}
		byAssignee[key].Count++
	}

	stats := &TaskStats{ByStatus: byStatus, Total: len(tasks)}
	for _, v := range byAssignee {
		stats.ByAssignee = append(stats.ByAssignee, *v)
	}
	if stats.ByAssignee == nil {
		stats.ByAssignee = []AssigneeStat{}
	}

	return stats, nil
}

// GetByID returns a task after verifying it exists.
// Exported so the stats endpoint and tests can use it.
func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	return s.tasks.GetByID(ctx, id)
}

// ProjectExists checks project exists (used by handlers for 404 before task ops).
func (s *TaskService) ProjectExists(ctx context.Context, projectID uuid.UUID) error {
	_, err := s.projects.GetByID(ctx, projectID)
	return err
}

