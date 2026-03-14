package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"tars/backend/internal/ws"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Agent represents the DB record for an agent.
type Agent struct {
	ID           string
	RepoID       string
	TaskID       *string
	Name         string
	Persona      *string
	Model        string
	SystemPrompt *string
	Status       string
	WorktreePath *string
	Branch       *string
	PID          *int
	StartedAt    *time.Time
	StoppedAt    *time.Time
}

// Manager maintains a registry of running agent processes.
type Manager struct {
	mu      sync.RWMutex
	agents  map[string]*Process // agentID → process
	hub     *ws.Hub
	db      *pgxpool.Pool
}

// NewManager creates a Manager.
func NewManager(db *pgxpool.Pool, hub *ws.Hub) *Manager {
	return &Manager{
		agents: make(map[string]*Process),
		hub:    hub,
		db:     db,
	}
}

// Spawn starts a Claude Code CLI process for the given agent.
// It updates the agent's status in the DB and broadcasts a WS event.
func (m *Manager) Spawn(ctx context.Context, a Agent) (*Process, error) {
	if a.WorktreePath == nil {
		return nil, fmt.Errorf("agent %s has no worktree_path", a.ID)
	}

	model := a.Model
	if model == "" {
		model = "claude-opus-4-5"
	}

	var systemPrompt string
	if a.SystemPrompt != nil {
		systemPrompt = *a.SystemPrompt
	}

	p, err := Spawn(ctx, SpawnConfig{
		AgentID:      a.ID,
		WorktreePath: *a.WorktreePath,
		Model:        model,
		SystemPrompt: systemPrompt,
		OnOutput: func(line LogLine) {
			m.broadcastOutput(a.ID, a.RepoID, line)
		},
		OnExit: func(exitCode int) {
			m.handleExit(a.ID, a.RepoID, exitCode)
		},
	})
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.agents[a.ID] = p
	m.mu.Unlock()

	// Update DB status and PID.
	pid := p.PID()
	now := time.Now()
	m.db.Exec(ctx,
		`UPDATE agents SET status='running', pid=$2, started_at=$3, updated_at=NOW() WHERE id=$1`,
		a.ID, pid, now,
	)

	// Broadcast agent.status event.
	m.broadcastStatus(a.RepoID, a.ID, "running", nil)

	return p, nil
}

// Stop sends SIGTERM to the agent process (then SIGKILL after 5s).
func (m *Manager) Stop(ctx context.Context, agentID, repoID string) error {
	m.mu.RLock()
	p, ok := m.agents[agentID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("agent %s is not running", agentID)
	}

	p.Stop()

	// DB + broadcast handled by OnExit callback.
	return nil
}

// Get returns the Process for a running agent.
func (m *Manager) Get(agentID string) (*Process, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.agents[agentID]
	return p, ok
}

// SendInput writes text to the agent's stdin.
func (m *Manager) SendInput(agentID, text string) error {
	m.mu.RLock()
	p, ok := m.agents[agentID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("agent %s is not running", agentID)
	}
	return p.SendInput(text)
}

// Logs returns log lines for an agent from its in-memory LogStore.
func (m *Manager) Logs(agentID string, limit, offset int) ([]LogLine, int64) {
	m.mu.RLock()
	p, ok := m.agents[agentID]
	m.mu.RUnlock()
	if !ok {
		return []LogLine{}, 0
	}
	return p.LogStore.Get(limit, offset)
}

// handleExit is called by the process goroutine when the subprocess exits.
func (m *Manager) handleExit(agentID, repoID string, exitCode int) {
	m.mu.Lock()
	delete(m.agents, agentID)
	m.mu.Unlock()

	status := "stopped"
	if exitCode != 0 {
		status = "crashed"
	}

	ctx := context.Background()
	now := time.Now()
	m.db.Exec(ctx,
		`UPDATE agents SET status=$2, stopped_at=$3, pid=NULL, updated_at=NOW() WHERE id=$1`,
		agentID, status, now,
	)

	m.broadcastStatus(repoID, agentID, status, &exitCode)
	log.Printf("agent %s exited with status=%s code=%d", agentID, status, exitCode)
}

// broadcastOutput sends an agent.output WS event.
func (m *Manager) broadcastOutput(agentID, repoID string, line LogLine) {
	if m.hub == nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"agent_id": agentID,
		"seq":      line.Seq,
		"ts":       line.Ts.Format(time.RFC3339Nano),
		"stream":   line.Stream,
		"text":     line.Text,
	})
	m.hub.Broadcast("agent:"+agentID, ws.Envelope{
		Type:    "agent.output",
		Channel: "agent:" + agentID,
		Payload: payload,
	})
}

// broadcastStatus sends an agent.status WS event to the repo channel.
func (m *Manager) broadcastStatus(repoID, agentID, status string, exitCode *int) {
	if m.hub == nil {
		return
	}
	p := map[string]any{
		"agent_id": agentID,
		"status":   status,
		"ts":       time.Now().UTC().Format(time.RFC3339),
	}
	if exitCode != nil {
		p["exit_code"] = *exitCode
	} else {
		p["exit_code"] = nil
	}
	payload, _ := json.Marshal(p)
	m.hub.Broadcast("repo:"+repoID, ws.Envelope{
		Type:    "agent.status",
		Channel: "repo:" + repoID,
		Payload: payload,
	})
}
