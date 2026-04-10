package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/handler"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/middleware"
)

func New(
	jwtSecret string,
	auth *handler.AuthHandler,
	project *handler.ProjectHandler,
	task *handler.TaskHandler,
	sse *handler.SSEHandler,
) http.Handler {
	r := chi.NewRouter()

	// ── Global middleware ────────────────────────────────────────────────────
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.Recoverer) // catch panics, return 500

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// ── Public routes (no auth) ──────────────────────────────────────────────
	r.Post("/auth/register", auth.Register)
	r.Post("/auth/login", auth.Login)

	// Health check — useful for Docker healthcheck
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// ── Protected routes (JWT required) ─────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))

		// Projects
		r.Get("/projects", project.List)
		r.Post("/projects", project.Create)
		r.Get("/projects/{id}", project.GetByID)
		r.Patch("/projects/{id}", project.Update)
		r.Delete("/projects/{id}", project.Delete)

		// Tasks nested under a project
		r.Get("/projects/{id}/tasks", task.ListByProject)
		r.Post("/projects/{id}/tasks", task.Create)

		// Bonus: project stats
		r.Get("/projects/{id}/stats", task.GetStats)

		// Bonus: SSE — task list changes (Authorization header via fetch-based client)
		r.Get("/projects/{id}/events", sse.ProjectTaskEvents)

		// Tasks by their own ID
		r.Patch("/tasks/{id}", task.Update)
		r.Delete("/tasks/{id}", task.Delete)
	})

	return r
}
