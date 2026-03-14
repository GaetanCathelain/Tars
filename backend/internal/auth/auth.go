package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const cookieName = "tars_session"

type contextKey string

const userIDKey contextKey = "userID"

// Config holds auth configuration.
type Config struct {
	SessionSecret []byte
	CookieSecure  bool
}

// Manager handles session creation and validation.
type Manager struct {
	cfg Config
}

// New creates a new auth Manager.
func New(cfg Config) *Manager {
	return &Manager{cfg: cfg}
}

// sessionPayload is the signed cookie payload.
type sessionPayload struct {
	UserID    string `json:"u"`
	IssuedAt  int64  `json:"i"`
}

// CreateSession sets a signed session cookie on the response.
func (m *Manager) CreateSession(w http.ResponseWriter, userID string) error {
	payload := sessionPayload{
		UserID:   userID,
		IssuedAt: time.Now().Unix(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal session payload: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(data)
	sig := m.sign(encoded)
	value := encoded + "." + sig

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
	return nil
}

// ValidateSession reads and validates the session cookie, returning the user ID.
func (m *Manager) ValidateSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("no session cookie")
	}

	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed session cookie")
	}

	encoded, sig := parts[0], parts[1]
	expectedSig := m.sign(encoded)
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return "", fmt.Errorf("invalid session signature")
	}

	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode session payload: %w", err)
	}

	var payload sessionPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", fmt.Errorf("unmarshal session payload: %w", err)
	}

	// Sessions expire after 7 days.
	if time.Since(time.Unix(payload.IssuedAt, 0)) > 7*24*time.Hour {
		return "", fmt.Errorf("session expired")
	}

	return payload.UserID, nil
}

// ClearSession expires the session cookie.
func (m *Manager) ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// RequireAuth is middleware that enforces authentication.
func (m *Manager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := m.ValidateSession(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID retrieves the authenticated user ID from the request context.
// Returns empty string if not authenticated.
func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

// GenerateState generates a random CSRF state token.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (m *Manager) sign(data string) string {
	mac := hmac.New(sha256.New, m.cfg.SessionSecret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
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
