package presence

import (
	"encoding/json"
	"sync"
	"time"

	"tars/backend/internal/ws"
)

const expireAfter = 90 * time.Second
const pruneInterval = 60 * time.Second

// UserPresence holds the current presence state for one user.
type UserPresence struct {
	UserID         string    `json:"user_id"`
	Login          string    `json:"login"`
	AvatarURL      string    `json:"avatar_url"`
	RepoID         string    `json:"repo_id"`
	ViewingAgentID *string   `json:"viewing_agent_id"`
	LastSeen       time.Time `json:"last_seen"`
}

// Snapshot is the full presence state for a repo (matches WS protocol).
type Snapshot struct {
	RepoID string         `json:"repo_id"`
	Users  []UserPresence `json:"users"`
}

// Tracker maintains in-memory user presence.
type Tracker struct {
	mu    sync.RWMutex
	users map[string]*UserPresence // userID → presence
	hub   *ws.Hub
}

// New creates a Tracker and starts the expiry goroutine.
func New(hub *ws.Hub) *Tracker {
	t := &Tracker{
		users: make(map[string]*UserPresence),
		hub:   hub,
	}
	go t.pruneLoop()
	return t
}

// Update upserts a user's presence and broadcasts a snapshot to their repo.
func (t *Tracker) Update(p UserPresence) {
	p.LastSeen = time.Now()

	t.mu.Lock()
	t.users[p.UserID] = &p
	repoID := p.RepoID
	t.mu.Unlock()

	t.broadcastSnapshot(repoID)
}

// Remove deletes a user's presence on disconnect and broadcasts an update.
func (t *Tracker) Remove(userID string) {
	t.mu.Lock()
	existing, ok := t.users[userID]
	if !ok {
		t.mu.Unlock()
		return
	}
	repoID := existing.RepoID
	delete(t.users, userID)
	t.mu.Unlock()

	t.broadcastSnapshot(repoID)
}

// GetByRepo returns all present users in a repo.
func (t *Tracker) GetByRepo(repoID string) []UserPresence {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []UserPresence
	for _, p := range t.users {
		if p.RepoID == repoID {
			result = append(result, *p)
		}
	}
	return result
}

// Snapshot builds the full snapshot payload for a repo.
func (t *Tracker) Snapshot(repoID string) Snapshot {
	users := t.GetByRepo(repoID)
	if users == nil {
		users = []UserPresence{}
	}
	return Snapshot{RepoID: repoID, Users: users}
}

// broadcastSnapshot sends presence.snapshot to all repo channel subscribers.
func (t *Tracker) broadcastSnapshot(repoID string) {
	if t.hub == nil {
		return
	}
	snap := t.Snapshot(repoID)
	payload, _ := json.Marshal(snap)
	t.hub.Broadcast("repo:"+repoID, ws.Envelope{
		Type:    "presence.snapshot",
		Channel: "repo:" + repoID,
		Payload: payload,
	})
}

// pruneLoop removes users who haven't been seen in expireAfter.
func (t *Tracker) pruneLoop() {
	ticker := time.NewTicker(pruneInterval)
	defer ticker.Stop()
	for range ticker.C {
		t.prune()
	}
}

func (t *Tracker) prune() {
	cutoff := time.Now().Add(-expireAfter)
	var stale []string

	t.mu.RLock()
	for uid, p := range t.users {
		if p.LastSeen.Before(cutoff) {
			stale = append(stale, uid)
		}
	}
	t.mu.RUnlock()

	for _, uid := range stale {
		t.Remove(uid)
	}
}
