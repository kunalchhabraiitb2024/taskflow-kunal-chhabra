package handler

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// ErrorJSON writes a simple { "error": "..." } response.
func ErrorJSON(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

// ValidationErrorJSON writes a 400 with field-level errors.
// { "error": "validation failed", "fields": { "email": "is required" } }
func ValidationErrorJSON(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusBadRequest, map[string]any{
		"error":  "validation failed",
		"fields": fields,
	})
}
