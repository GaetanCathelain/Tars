package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/GaetanCathelain/Tars/internal/model"
	wshub "github.com/GaetanCathelain/Tars/internal/ws"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultTimeout = 15 * time.Minute

// Manager handles the lifecycle of worker sessions (PTY-based Claude processes).
type Manager struct {
	db     *pgxpool.Pool
	hub    *wshub.Hub
	mu     sync.Mutex
	active map[uuid.UUID]*Session
}

// Session represents a running worker process.
type Session struct {
	ID        uuid.UUID
	TaskID    uuid.UUID
	Command   string
	Process   *os.Process
	PTY       *os.File
	Status    string
	StartedAt time.Time
	Cancel    context.CancelFunc
}

// NewManager creates a new worker manager.
func NewManager(db *pgxpool.Pool, hub *wshub.Hub) *Manager {
	return &Manager{
		db:     db,
		hub:    hub,
		active: make(map[uuid.UUID]*Session),
	}
}

// SpawnWorker creates a PTY, spawns claude, starts output capture, and manages lifecycle.
func (m *Manager) SpawnWorker(ctx context.Context, taskID uuid.UUID, messageID *uuid.UUID, prompt string) (*model.WorkerSession, error) {
	sessionID := uuid.New()
	now := time.Now()
	command := fmt.Sprintf("claude %q", prompt)

	// Insert DB row
	wSession := &model.WorkerSession{
		ID:        sessionID,
		TaskID:    taskID,
		MessageID: messageID,
		Status:    "running",
		Command:   command,
		StartedAt: now,
	}

	_, err := m.db.Exec(ctx,
		`INSERT INTO worker_sessions (id, task_id, message_id, status, command, started_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		wSession.ID, wSession.TaskID, wSession.MessageID, wSession.Status, wSession.Command, wSession.StartedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert worker_session: %w", err)
	}

	// Create cancellable context with timeout
	workerCtx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	// Spawn PTY process
	ptmx, cmd, err := spawnClaude(prompt)
	if err != nil {
		cancel()
		m.updateSessionStatus(sessionID, "failed", nil)
		return nil, fmt.Errorf("spawn claude: %w", err)
	}

	session := &Session{
		ID:        sessionID,
		TaskID:    taskID,
		Command:   command,
		Process:   cmd.Process,
		PTY:       ptmx,
		Status:    "running",
		StartedAt: now,
		Cancel:    cancel,
	}

	m.mu.Lock()
	m.active[sessionID] = session
	m.mu.Unlock()

	// Broadcast worker_start
	m.hub.BroadcastToTask(taskID, &wshub.OutgoingMessage{
		Type:      "worker_start",
		TaskID:    taskID,
		SessionID: sessionID,
		Session:   wSession,
		Status:    "running",
	})

	// Start output capture goroutine
	outputDone := make(chan struct{})
	go captureOutput(workerCtx, m.db, m.hub, sessionID, taskID, ptmx, outputDone)

	// Start process wait goroutine
	go m.waitForExit(workerCtx, cancel, cmd, session, outputDone)

	return wSession, nil
}

// KillWorker terminates a running worker session.
func (m *Manager) KillWorker(sessionID uuid.UUID) error {
	m.mu.Lock()
	session, ok := m.active[sessionID]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("session %s not found or not running", sessionID)
	}

	// Cancel context triggers cleanup
	session.Cancel()

	// Also send SIGKILL to be sure
	if session.Process != nil {
		if err := session.Process.Kill(); err != nil {
			slog.Warn("kill worker process", "session_id", sessionID, "error", err)
		}
	}

	return nil
}

// GetSession returns an active session by ID.
func (m *Manager) GetSession(sessionID uuid.UUID) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active[sessionID]
}

// ActiveSessions returns all currently running sessions.
func (m *Manager) ActiveSessions() []*Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessions := make([]*Session, 0, len(m.active))
	for _, s := range m.active {
		sessions = append(sessions, s)
	}
	return sessions
}

// waitForExit waits for the claude process to exit, then cleans up.
func (m *Manager) waitForExit(ctx context.Context, cancel context.CancelFunc, cmd interface{ Wait() error }, session *Session, outputDone chan struct{}) {
	defer cancel()

	// Wait for process or context cancellation
	exitCh := make(chan error, 1)
	go func() {
		exitCh <- cmd.Wait()
	}()

	var exitErr error
	select {
	case exitErr = <-exitCh:
		// Process exited naturally
	case <-ctx.Done():
		// Timeout or cancelled — kill process
		if session.Process != nil {
			session.Process.Kill()
		}
		exitErr = <-exitCh // Wait for it to actually die
	}

	// Close PTY to signal EOF to capture goroutine
	session.PTY.Close()

	// Wait for output capture to finish flushing
	<-outputDone

	// Determine exit code and status
	var exitCode int
	status := "completed"
	if exitErr != nil {
		status = "failed"
		// Try to extract exit code
		if exitError, ok := exitErr.(*os.PathError); ok {
			slog.Warn("worker exit path error", "error", exitError)
			exitCode = 1
		} else {
			exitCode = 1
		}
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			status = "timeout"
			exitCode = -1
		}
	}

	slog.Info("worker exited",
		"session_id", session.ID,
		"task_id", session.TaskID,
		"status", status,
		"exit_code", exitCode,
	)

	// Update DB
	m.updateSessionFinished(session.ID, status, exitCode)

	// Remove from active map
	m.mu.Lock()
	delete(m.active, session.ID)
	m.mu.Unlock()

	// Broadcast worker_end
	m.hub.BroadcastToTask(session.TaskID, &wshub.OutgoingMessage{
		Type:      "worker_end",
		TaskID:    session.TaskID,
		SessionID: session.ID,
		ExitCode:  &exitCode,
		Status:    status,
	})
}

func (m *Manager) updateSessionStatus(sessionID uuid.UUID, status string, exitCode *int) {
	_, err := m.db.Exec(context.Background(),
		`UPDATE worker_sessions SET status = $1, exit_code = $2 WHERE id = $3`,
		status, exitCode, sessionID,
	)
	if err != nil {
		slog.Error("update worker_session status", "session_id", sessionID, "error", err)
	}
}

func (m *Manager) updateSessionFinished(sessionID uuid.UUID, status string, exitCode int) {
	now := time.Now()
	_, err := m.db.Exec(context.Background(),
		`UPDATE worker_sessions SET status = $1, exit_code = $2, finished_at = $3 WHERE id = $4`,
		status, exitCode, now, sessionID,
	)
	if err != nil {
		slog.Error("update worker_session finished", "session_id", sessionID, "error", err)
	}
}
