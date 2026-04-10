package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type userIDKeyType struct{}

var userIDContextKey = userIDKeyType{}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// Auth returns a middleware that validates the Bearer JWT.
// Protected routes are wrapped with this middleware.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID from the request context.
// Panics if called outside an authenticated route (programming error, not user error).
func UserIDFromContext(ctx context.Context) uuid.UUID {
	id, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	if !ok {
		panic("UserIDFromContext called on unauthenticated request")
	}
	return id
}
