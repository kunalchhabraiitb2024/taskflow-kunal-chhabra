package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/middleware"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/model"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/repository"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/service"
)

type TaskHandler struct {
	tasks *service.TaskService
}

func NewTaskHandler(tasks *service.TaskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

func (h *TaskHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid project id")
		return
	}

	// Optional filters
	var statusFilter *model.TaskStatus
	if s := r.URL.Query().Get("status"); s != "" {
		ts := model.TaskStatus(s)
		if !ts.Valid() {
			ErrorJSON(w, http.StatusBadRequest, "invalid status filter")
			return
		}
		statusFilter = &ts
	}

	var assigneeFilter *uuid.UUID
	if a := r.URL.Query().Get("assignee"); a != "" {
		id, err := uuid.Parse(a)
		if err != nil {
			ErrorJSON(w, http.StatusBadRequest, "invalid assignee filter")
			return
		}
		assigneeFilter = &id
	}

	result, err := h.tasks.ListByProject(r.Context(), projectID, statusFilter, assigneeFilter, parsePagination(r))
	if errors.Is(err, repository.ErrNotFound) {
		ErrorJSON(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	JSON(w, http.StatusOK, result)
}

type createTaskRequest struct {
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	Status      string       `json:"status"`
	Priority    string       `json:"priority"`
	AssigneeID  *string      `json:"assignee_id"`
	DueDate     *string      `json:"due_date"` // "YYYY-MM-DD"
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projectID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := map[string]string{}
	if strings.TrimSpace(req.Title) == "" {
		fields["title"] = "is required"
	}
	if req.Status != "" && !model.TaskStatus(req.Status).Valid() {
		fields["status"] = "must be todo, in_progress, or done"
	}
	if req.Priority != "" && !model.TaskPriority(req.Priority).Valid() {
		fields["priority"] = "must be low, medium, or high"
	}
	if len(fields) > 0 {
		ValidationErrorJSON(w, fields)
		return
	}

	t := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatus(req.Status),
		Priority:    model.TaskPriority(req.Priority),
	}
	if req.AssigneeID != nil {
		id, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			ValidationErrorJSON(w, map[string]string{"assignee_id": "invalid UUID"})
			return
		}
		t.AssigneeID = &id
	}
	if req.DueDate != nil {
		due, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			ValidationErrorJSON(w, map[string]string{"due_date": "must be YYYY-MM-DD"})
			return
		}
		t.DueDate = &due
	}

	created, err := h.tasks.Create(r.Context(), userID, projectID, t)
	if errors.Is(err, repository.ErrNotFound) {
		ErrorJSON(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	JSON(w, http.StatusCreated, created)
}

type updateTaskRequest struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	Status       *string `json:"status"`
	Priority     *string `json:"priority"`
	AssigneeID   *string `json:"assignee_id"` // empty string = unassign
	DueDate      *string `json:"due_date"`    // empty string = clear
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := map[string]string{}
	if req.Status != nil && !model.TaskStatus(*req.Status).Valid() {
		fields["status"] = "must be todo, in_progress, or done"
	}
	if req.Priority != nil && !model.TaskPriority(*req.Priority).Valid() {
		fields["priority"] = "must be low, medium, or high"
	}
	if len(fields) > 0 {
		ValidationErrorJSON(w, fields)
		return
	}

	u := repository.TaskUpdate{}
	u.Title = req.Title
	u.Description = req.Description
	if req.Status != nil {
		s := model.TaskStatus(*req.Status)
		u.Status = &s
	}
	if req.Priority != nil {
		p := model.TaskPriority(*req.Priority)
		u.Priority = &p
	}
	if req.AssigneeID != nil {
		if *req.AssigneeID == "" {
			// Empty string = explicitly unassign
			u.ClearAssigneeID = true
		} else {
			aid, err := uuid.Parse(*req.AssigneeID)
			if err != nil {
				ValidationErrorJSON(w, map[string]string{"assignee_id": "invalid UUID"})
				return
			}
			u.AssigneeID = &aid
		}
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			u.ClearDueDate = true
		} else {
			due, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				ValidationErrorJSON(w, map[string]string{"due_date": "must be YYYY-MM-DD"})
				return
			}
			u.DueDate = &due
		}
	}

	task, err := h.tasks.Update(r.Context(), id, u)
	if errors.Is(err, repository.ErrNotFound) {
		ErrorJSON(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	JSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid task id")
		return
	}

	err = h.tasks.Delete(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		ErrorJSON(w, http.StatusNotFound, "not found")
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		ErrorJSON(w, http.StatusForbidden, "forbidden")
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to delete task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid project id")
		return
	}
	stats, err := h.tasks.GetStats(r.Context(), projectID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	JSON(w, http.StatusOK, stats)
}
