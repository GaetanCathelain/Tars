package api

import (
	"context"
	"fmt"
	"strings"

	"tars/backend/internal/git"
)

// gitDiff delegates to the git package's Diff function.
func gitDiff(repoPath, baseRef, headRef, format, agentID string) (*git.DiffResult, error) {
	return git.Diff(repoPath, baseRef, headRef, format, agentID)
}

// gitMerge merges agentBranch into targetBranch using the given strategy.
// Returns the resulting commit SHA on success.
func gitMerge(ctx context.Context, repoPath, agentBranch, targetBranch, strategy, commitMessage string) (string, error) {
	// Checkout target branch.
	if out, err := runGitInDir(ctx, repoPath, "checkout", targetBranch); err != nil {
		return "", fmt.Errorf("checkout %s: %s", targetBranch, out)
	}

	var mergeArgs []string
	switch strategy {
	case "squash":
		mergeArgs = []string{"merge", "--squash", agentBranch}
	case "rebase":
		// Rebase agent branch onto target, then fast-forward target.
		if out, err := runGitInDir(ctx, repoPath, "checkout", agentBranch); err != nil {
			return "", fmt.Errorf("checkout agent branch: %s", out)
		}
		if out, err := runGitInDir(ctx, repoPath, "rebase", targetBranch); err != nil {
			return "", fmt.Errorf("rebase: %s", out)
		}
		if out, err := runGitInDir(ctx, repoPath, "checkout", targetBranch); err != nil {
			return "", fmt.Errorf("checkout target: %s", out)
		}
		mergeArgs = []string{"merge", "--ff-only", agentBranch}
	default: // "merge"
		mergeArgs = []string{"merge", "--no-ff", agentBranch}
	}

	if out, err := runGitInDir(ctx, repoPath, mergeArgs...); err != nil {
		return "", fmt.Errorf("merge conflict: %s", out)
	}

	// For squash, we need an explicit commit.
	if strategy == "squash" {
		msg := commitMessage
		if msg == "" {
			msg = "chore: squash merge " + agentBranch
		}
		if out, err := runGitInDir(ctx, repoPath, "commit", "-m", msg); err != nil {
			return "", fmt.Errorf("commit squash: %s", out)
		}
	}

	// Get the resulting HEAD SHA.
	sha, err := runGitInDir(ctx, repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("rev-parse HEAD: %w", err)
	}

	return strings.TrimSpace(sha), nil
}

// runGitInDir runs a git command in dir using the git package's runner.
func runGitInDir(ctx context.Context, dir string, args ...string) (string, error) {
	return git.RunGit(ctx, dir, args...)
}
