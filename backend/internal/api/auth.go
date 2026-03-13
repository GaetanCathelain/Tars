package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/GaetanCathelain/Tars/internal/db"
	"github.com/GaetanCathelain/Tars/internal/models"
)

type AuthHandler struct {
	queries       *db.Queries
	clientID      string
	clientSecret  string
	redirectURI   string
	sessionSecret string
	frontendURL   string
}

func NewAuthHandler(queries *db.Queries, clientID, clientSecret, redirectURI, sessionSecret, frontendURL string) *AuthHandler {
	return &AuthHandler{
		queries:       queries,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   redirectURI,
		sessionSecret: sessionSecret,
		frontendURL:   frontendURL,
	}
}

// GitHubLogin redirects to GitHub OAuth authorize URL.
func (h *AuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user,user:email",
		h.clientID, h.redirectURI,
	)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GitHubCallback handles the OAuth callback.
func (h *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing code parameter")
		return
	}

	// Exchange code for access token.
	accessToken, err := h.exchangeCode(code)
	if err != nil {
		log.Printf("github code exchange failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to authenticate with github")
		return
	}

	// Fetch GitHub user info.
	ghUser, err := h.fetchGitHubUser(accessToken)
	if err != nil {
		log.Printf("github user fetch failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch github user")
		return
	}

	// Upsert user.
	user := &models.User{
		GitHubID:    ghUser.ID,
		Username:    ghUser.Login,
		Email:       ghUser.Email,
		AvatarURL:   ghUser.AvatarURL,
		AccessToken: accessToken,
	}
	if err := h.queries.UpsertUser(r.Context(), user); err != nil {
		log.Printf("upsert user failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Create session.
	token := h.generateSessionToken()
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	if _, err := h.queries.CreateSession(r.Context(), user.ID, token, expiresAt); err != nil {
		log.Printf("create session failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Set cookie and redirect to frontend.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, h.frontendURL, http.StatusTemporaryRedirect)
}

// Logout invalidates the session.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		h.queries.DeleteSession(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// Me returns the current authenticated user.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// --- GitHub API helpers ---

type githubUser struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Email     *string `json:"email"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *AuthHandler) exchangeCode(code string) (string, error) {
	payload := fmt.Sprintf(
		`{"client_id":"%s","client_secret":"%s","code":"%s"}`,
		h.clientID, h.clientSecret, code,
	)

	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token",
		io.NopCloser(jsonReader(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("github error: %s", result.Error)
	}
	return result.AccessToken, nil
}

func (h *AuthHandler) fetchGitHubUser(accessToken string) (*githubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var user githubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &user, nil
}

func (h *AuthHandler) generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	mac := hmac.New(sha256.New, []byte(h.sessionSecret))
	mac.Write(b)
	return hex.EncodeToString(mac.Sum(nil))
}
