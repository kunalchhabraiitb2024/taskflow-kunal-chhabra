package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/config"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/database"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/handler"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/realtime"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/repository"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/router"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/service"
)

func main() {
	// ── Structured logger ────────────────────────────────────────────────────
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// ── Load config ──────────────────────────────────────────────────────────
	cfg := config.Load()

	// ── Connect to database ──────────────────────────────────────────────────
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to database")

	// ── Run migrations ───────────────────────────────────────────────────────
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations up to date")

	// ── Run seed (idempotent) ────────────────────────────────────────────────
	if err := runSeed(ctx, pool); err != nil {
		// Seed failure is non-fatal — log and continue
		slog.Warn("seed failed (non-fatal)", "error", err)
	} else {
		slog.Info("seed data ready")
	}

	// ── Dependency injection ─────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)

	broker := realtime.NewBroker()

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.BcryptCost)
	projectSvc := service.NewProjectService(projectRepo, taskRepo, broker)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, broker)

	authHandler := handler.NewAuthHandler(authSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	sseHandler := handler.NewSSEHandler(broker)

	// ── Build router ─────────────────────────────────────────────────────────
	r := router.New(cfg.JWTSecret, authHandler, projectHandler, taskHandler, sseHandler)

	// ── Start HTTP server ────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // SSE / long-lived responses
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// ── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "error", err)
	}
	slog.Info("server stopped")
}

func runMigrations(databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

func runSeed(ctx context.Context, pool *pgxpool.Pool) error {
	data, err := os.ReadFile("seed/seed.sql")
	if err != nil {
		return fmt.Errorf("read seed file: %w", err)
	}

	// pgxpool.Pool.Exec uses the extended query protocol which doesn't support
	// multi-statement SQL. Acquire a raw connection and use pgconn.Exec which
	// uses the simple query protocol and handles multiple statements fine.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection for seed: %w", err)
	}
	defer conn.Release()

	mrr := conn.Conn().PgConn().Exec(ctx, string(data))
	_, err = mrr.ReadAll()
	if err != nil {
		return fmt.Errorf("exec seed: %w", err)
	}
	return mrr.Close()
}
