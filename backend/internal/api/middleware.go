package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/GaetanCathelain/Tars/internal/db"
	"github.com/GaetanCathelain/Tars/internal/models"
	"github.com/google/uuid"
)

// RequestID injects a unique request ID into the context and response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), models.ContextKeyRequestID, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger logs request method, path, status, and duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, ww.status, time.Since(start).Round(time.Millisecond))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Recovery recovers from panics and returns 500.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Auth extracts the session cookie, validates it, and injects the user into context.
func Auth(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			session, err := queries.GetSessionByToken(r.Context(), cookie.Value)
			if err != nil {
				log.Printf("session lookup error: %v", err)
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			if session == nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired session")
				return
			}

			user, err := queries.GetUserByID(r.Context(), session.UserID)
			if err != nil || user == nil {
				writeError(w, http.StatusUnauthorized, "user not found")
				return
			}

			ctx := context.WithValue(r.Context(), models.ContextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext extracts the authenticated user from request context.
func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(models.ContextKeyUser).(*models.User)
	return u
}
