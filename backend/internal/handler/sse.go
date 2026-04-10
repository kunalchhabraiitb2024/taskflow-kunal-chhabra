package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/middleware"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/realtime"
)

type SSEHandler struct {
	broker *realtime.Broker
}

func NewSSEHandler(b *realtime.Broker) *SSEHandler {
	return &SSEHandler{broker: b}
}

// ProjectTaskEvents streams Server-Sent Events when tasks in the project change.
// Requires JWT (Authorization: Bearer) — use fetch-based EventSource clients that send headers.
func (h *SSEHandler) ProjectTaskEvents(w http.ResponseWriter, r *http.Request) {
	if h.broker == nil {
		http.Error(w, `{"error":"realtime unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	_ = middleware.UserIDFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid project id"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"streaming unsupported"}`, http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	ch := h.broker.Subscribe(ctx, projectID)

	// Comment line so clients know the stream is alive
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "event: tasks\ndata: changed\n\n")
			flusher.Flush()
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
