package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserURL      = "https://api.github.com/user"
	oauthScope         = "read:user user:email"
	stateCookieName    = "tars_oauth_state"
)

// GitHubConfig holds OAuth app credentials.
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
}

// GitHubUser is the data returned from the GitHub user API.
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// GitHub handles the GitHub OAuth flow.
type GitHub struct {
	cfg     GitHubConfig
	session *Manager
}

// NewGitHub creates a GitHub OAuth handler.
func NewGitHub(cfg GitHubConfig, session *Manager) *GitHub {
	return &GitHub{cfg: cfg, session: session}
}

// HandleLogin redirects the user to GitHub's OAuth authorize page.
func (g *GitHub) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := GenerateState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate state", nil)
		return
	}

	// Store state in a short-lived cookie for CSRF validation in the callback.
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   g.session.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	params := url.Values{
		"client_id": {g.cfg.ClientID},
		"scope":     {oauthScope},
		"state":     {state},
	}
	http.Redirect(w, r, githubAuthorizeURL+"?"+params.Encode(), http.StatusFound)
}

// HandleCallback processes the GitHub OAuth callback.
// On success, sets the session cookie and redirects to /dashboard.
// On failure, redirects to /login?error=<reason>.
func (g *GitHub) HandleCallback(w http.ResponseWriter, r *http.Request, upsertUser func(ctx context.Context, u GitHubUser) (string, error)) {
	// Validate CSRF state.
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || stateCookie.Value == "" {
		http.Redirect(w, r, "/login?error=missing_state", http.StatusFound)
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		http.Redirect(w, r, "/login?error=invalid_state", http.StatusFound)
		return
	}

	// Clear the state cookie.
	http.SetCookie(w, &http.Cookie{
		Name:   stateCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/login?error=missing_code", http.StatusFound)
		return
	}

	// Exchange code for access token.
	token, err := g.exchangeCode(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, "/login?error=token_exchange_failed", http.StatusFound)
		return
	}

	// Fetch GitHub user profile.
	ghUser, err := g.fetchUser(r.Context(), token)
	if err != nil {
		http.Redirect(w, r, "/login?error=user_fetch_failed", http.StatusFound)
		return
	}

	// Upsert user in DB and get TARS user ID.
	userID, err := upsertUser(r.Context(), *ghUser)
	if err != nil {
		http.Redirect(w, r, "/login?error=db_error", http.StatusFound)
		return
	}

	// Create session cookie.
	if err := g.session.CreateSession(w, userID); err != nil {
		http.Redirect(w, r, "/login?error=session_error", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (g *GitHub) exchangeCode(ctx context.Context, code string) (string, error) {
	body := url.Values{
		"client_id":     {g.cfg.ClientID},
		"client_secret": {g.cfg.ClientSecret},
		"code":          {code},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("github oauth error: %s", result.Error)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}

	return result.AccessToken, nil
}

func (g *GitHub) fetchUser(ctx context.Context, token string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github user api %d: %s", resp.StatusCode, body)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}
	return &user, nil
}
