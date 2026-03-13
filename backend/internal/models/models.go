package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	GitHubID    int64     `json:"github_id"`
	Username    string    `json:"username"`
	Email       *string   `json:"email,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	AccessToken string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Repo struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	LocalPath     *string    `json:"local_path,omitempty"`
	DefaultBranch string     `json:"default_branch"`
	AddedBy       *uuid.UUID `json:"added_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Task struct {
	ID          uuid.UUID  `json:"id"`
	RepoID      *uuid.UUID `json:"repo_id,omitempty"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	Priority    int        `json:"priority"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Agent struct {
	ID           uuid.UUID  `json:"id"`
	TaskID       *uuid.UUID `json:"task_id,omitempty"`
	RepoID       *uuid.UUID `json:"repo_id,omitempty"`
	Name         string     `json:"name"`
	Persona      *string    `json:"persona,omitempty"`
	WorktreePath *string    `json:"worktree_path,omitempty"`
	BranchName   *string    `json:"branch_name,omitempty"`
	PID          *int       `json:"pid,omitempty"`
	Status       string     `json:"status"`
	Model        string     `json:"model"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Message struct {
	ID        uuid.UUID  `json:"id"`
	AgentID   *uuid.UUID `json:"agent_id,omitempty"`
	TaskID    *uuid.UUID `json:"task_id,omitempty"`
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
}

type Event struct {
	ID        uuid.UUID        `json:"id"`
	TaskID    *uuid.UUID       `json:"task_id,omitempty"`
	AgentID   *uuid.UUID       `json:"agent_id,omitempty"`
	UserID    *uuid.UUID       `json:"user_id,omitempty"`
	Type      string           `json:"type"`
	Payload   *json.RawMessage `json:"payload,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// Context keys for request-scoped values.
type contextKey string

const (
	ContextKeyUser      contextKey = "user"
	ContextKeyRequestID contextKey = "request_id"
)
