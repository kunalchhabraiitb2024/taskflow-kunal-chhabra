package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/middleware"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/repository"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/service"
)

type ProjectHandler struct {
	projects *service.ProjectService
}

func NewProjectHandler(projects *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projects: projects}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	result, err := h.projects.List(r.Context(), userID, parsePagination(r))
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	JSON(w, http.StatusOK, result)
}

type createProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		ValidationErrorJSON(w, map[string]string{"name": "is required"})
		return
	}

	project, err := h.projects.Create(r.Context(), userID, req.Name, req.Description)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to create project")
		return
	}
	JSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.projects.GetByIDWithTasks(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		ErrorJSON(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to get project")
		return
	}
	JSON(w, http.StatusOK, project)
}

type updateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := h.projects.Update(r.Context(), id, userID, req.Name, req.Description)
	if errors.Is(err, repository.ErrNotFound) {
		ErrorJSON(w, http.StatusNotFound, "not found")
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		ErrorJSON(w, http.StatusForbidden, "forbidden")
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to update project")
		return
	}
	JSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid project id")
		return
	}

	err = h.projects.Delete(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		ErrorJSON(w, http.StatusNotFound, "not found")
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		ErrorJSON(w, http.StatusForbidden, "forbidden")
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseUUID is a shared helper for URL param parsing.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
