package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const cmdTimeout = 30 * time.Second

// Worktree represents a git worktree for an agent.
type Worktree struct {
	Path    string
	Branch  string
	AgentID string
	Head    string // current HEAD commit SHA
}

// WorktreeManager manages git worktrees for agent isolation.
type WorktreeManager struct{}

// NewWorktreeManager returns a WorktreeManager.
func NewWorktreeManager() *WorktreeManager {
	return &WorktreeManager{}
}

// worktreePath returns the canonical worktree path for an agent.
func worktreePath(repoPath, agentID string) string {
	return filepath.Join(repoPath, ".tars", "agents", agentID)
}

// branchName returns the git branch name for an agent.
func branchName(agentID string) string {
	return "tars/agent-" + agentID
}

// Create creates a new git worktree for the given agent.
// The worktree is created at <repoPath>/.tars/agents/<agentID> on branch tars/agent-<agentID>.
// Idempotent: if the worktree already exists, returns its current state.
func (m *WorktreeManager) Create(ctx context.Context, repoPath, agentID string) (*Worktree, error) {
	path := worktreePath(repoPath, agentID)
	branch := branchName(agentID)

	// Check if worktree already exists.
	existing, err := m.findWorktree(ctx, repoPath, path)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Create the worktree on a new branch.
	out, err := runGit(ctx, repoPath, "worktree", "add", "-b", branch, path)
	if err != nil {
		return nil, fmt.Errorf("git worktree add: %w\n%s", err, out)
	}

	head, _ := headSHA(ctx, path)
	return &Worktree{
		Path:    path,
		Branch:  branch,
		AgentID: agentID,
		Head:    head,
	}, nil
}

// Remove removes the worktree and deletes its branch.
func (m *WorktreeManager) Remove(ctx context.Context, repoPath, agentID string) error {
	path := worktreePath(repoPath, agentID)
	branch := branchName(agentID)

	if out, err := runGit(ctx, repoPath, "worktree", "remove", "--force", path); err != nil {
		// If the worktree doesn't exist, that's fine.
		if !strings.Contains(out, "is not a working tree") {
			return fmt.Errorf("git worktree remove: %w\n%s", err, out)
		}
	}

	// Delete the branch (best-effort).
	runGit(ctx, repoPath, "branch", "-D", branch) //nolint:errcheck

	return nil
}

// List returns all worktrees for a repo (excluding the main worktree).
func (m *WorktreeManager) List(ctx context.Context, repoPath string) ([]Worktree, error) {
	out, err := runGit(ctx, repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}

	return parseWorktreeList(out, repoPath), nil
}

// Prune cleans up stale worktree administrative files.
func (m *WorktreeManager) Prune(ctx context.Context, repoPath string) error {
	_, err := runGit(ctx, repoPath, "worktree", "prune")
	return err
}

// findWorktree looks up an existing worktree by path.
func (m *WorktreeManager) findWorktree(ctx context.Context, repoPath, path string) (*Worktree, error) {
	worktrees, err := m.List(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	for _, wt := range worktrees {
		if wt.Path == path {
			return &wt, nil
		}
	}
	return nil, fmt.Errorf("worktree not found: %s", path)
}

// parseWorktreeList parses `git worktree list --porcelain` output.
func parseWorktreeList(out, mainRepoPath string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" && current.Path != mainRepoPath {
				worktrees = append(worktrees, current)
			}
			current = Worktree{}
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Head = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}
	// Handle last entry.
	if current.Path != "" && current.Path != mainRepoPath {
		worktrees = append(worktrees, current)
	}

	// Extract agentID from branch name.
	for i, wt := range worktrees {
		if strings.HasPrefix(wt.Branch, "tars/agent-") {
			worktrees[i].AgentID = strings.TrimPrefix(wt.Branch, "tars/agent-")
		}
	}

	return worktrees
}

// headSHA returns the HEAD commit SHA for a repo/worktree path.
func headSHA(ctx context.Context, repoPath string) (string, error) {
	out, err := runGit(ctx, repoPath, "rev-parse", "HEAD")
	return strings.TrimSpace(out), err
}

// RunGit executes a git command in dir and returns combined stdout+stderr.
// Exported for use by other packages (e.g., api/git.go for merge operations).
func RunGit(ctx context.Context, dir string, args ...string) (string, error) {
	return runGit(ctx, dir, args...)
}

// runGit is the internal implementation.
func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	return buf.String(), err
}
