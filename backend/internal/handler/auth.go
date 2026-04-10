package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Client-side-style validation at the handler layer
	fields := map[string]string{}
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "is required"
	}
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	}
	if len(req.Password) < 6 {
		fields["password"] = "must be at least 6 characters"
	}
	if len(fields) > 0 {
		ValidationErrorJSON(w, fields)
		return
	}

	result, err := h.auth.Register(r.Context(), req.Name, req.Email, req.Password)
	if errors.Is(err, service.ErrEmailTaken) {
		ValidationErrorJSON(w, map[string]string{"email": "already in use"})
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "registration failed")
		return
	}

	JSON(w, http.StatusCreated, map[string]any{
		"token": result.Token,
		"user":  result.User.ToPublic(),
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := map[string]string{}
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	}
	if strings.TrimSpace(req.Password) == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		ValidationErrorJSON(w, fields)
		return
	}

	result, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		ErrorJSON(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "login failed")
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"token": result.Token,
		"user":  result.User.ToPublic(),
	})
}
