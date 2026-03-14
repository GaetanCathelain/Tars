package api

import (
	"encoding/json"
	"net/http"
	"strings"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// CORSMiddleware builds the CORS handler from a comma-separated list of allowed origins.
func CORSMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	origins := []string{}
	for _, o := range strings.Split(allowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://localhost:3000"}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// StandardMiddleware returns the ordered slice of middleware applied to all routes.
func StandardMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		chimiddleware.RealIP,
		chimiddleware.RequestID,
		chimiddleware.Logger,
		chimiddleware.Recoverer,
	}
}

// writeError writes a contract-compliant error JSON response.
func writeError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
			"details": details,
		},
	}
	data, _ := json.Marshal(body)
	w.Write(data)
}

// writeJSON writes a JSON response with status 200.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// notImplemented is a stub handler for routes not yet wired up.
func notImplemented(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "this endpoint is not yet implemented", nil)
}
