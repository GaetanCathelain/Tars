package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// DiffResult is the structured response for a git diff.
type DiffResult struct {
	AgentID string      `json:"agent_id"`
	BaseRef string      `json:"base_ref"`
	HeadRef string      `json:"head_ref"`
	Stats   DiffStats   `json:"stats"`
	Files   []FileDiff  `json:"files"`
}

// DiffStats holds aggregate diff statistics.
type DiffStats struct {
	FilesChanged int `json:"files_changed"`
	Insertions   int `json:"insertions"`
	Deletions    int `json:"deletions"`
}

// FileDiff holds diff data for a single file.
type FileDiff struct {
	Path      string `json:"path"`
	Status    string `json:"status"` // added, modified, deleted, renamed, copied
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch,omitempty"`
}

// Diff returns the diff of headRef against baseRef in repoPath.
// format: "unified" (default) or "stat".
func Diff(repoPath, baseRef, headRef, format, agentID string) (*DiffResult, error) {
	result := &DiffResult{
		AgentID: agentID,
		BaseRef: baseRef,
		HeadRef: headRef,
		Files:   []FileDiff{},
	}

	ctx := context.Background()

	// --- Pass 1: name-status to get per-file status (A/M/D/R/C) ---
	nsOut, err := runGit(ctx, repoPath, "diff", "--name-status", baseRef+"..."+headRef)
	if err != nil {
		return nil, fmt.Errorf("git diff --name-status: %w", err)
	}
	statusMap := parseNameStatus(nsOut)

	// --- Pass 2: numstat to get per-file line counts ---
	numOut, err := runGit(ctx, repoPath, "diff", "--numstat", baseRef+"..."+headRef)
	if err != nil {
		return nil, fmt.Errorf("git diff --numstat: %w", err)
	}
	fileMap := parseNumstat(numOut, statusMap)

	// --- Pass 3 (unified only): full patch ---
	if format != "stat" {
		patchOut, err := runGit(ctx, repoPath, "diff", "--unified=3", baseRef+"..."+headRef)
		if err != nil {
			return nil, fmt.Errorf("git diff --unified: %w", err)
		}
		applyPatches(fileMap, patchOut)
	}

	// Collect files and compute stats.
	for _, fd := range fileMap {
		result.Files = append(result.Files, fd)
		result.Stats.FilesChanged++
		result.Stats.Insertions += fd.Additions
		result.Stats.Deletions += fd.Deletions
	}

	return result, nil
}

// parseNameStatus parses `git diff --name-status` output.
// Returns map of path → status string.
func parseNameStatus(out string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		code := parts[0]
		path := parts[len(parts)-1]
		var status string
		switch {
		case strings.HasPrefix(code, "A"):
			status = "added"
		case strings.HasPrefix(code, "D"):
			status = "deleted"
		case strings.HasPrefix(code, "R"):
			status = "renamed"
		case strings.HasPrefix(code, "C"):
			status = "copied"
		default:
			status = "modified"
		}
		m[path] = status
	}
	return m
}

// parseNumstat parses `git diff --numstat` output.
func parseNumstat(out string, statusMap map[string]string) map[string]FileDiff {
	m := make(map[string]FileDiff)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		add, _ := strconv.Atoi(parts[0])
		del, _ := strconv.Atoi(parts[1])
		path := parts[2]
		status := statusMap[path]
		if status == "" {
			status = "modified"
		}
		m[path] = FileDiff{
			Path:      path,
			Status:    status,
			Additions: add,
			Deletions: del,
		}
	}
	return m
}

// applyPatches splits a unified diff by file and stores patches in fileMap.
func applyPatches(fileMap map[string]FileDiff, patchOut string) {
	sections := strings.Split(patchOut, "diff --git ")
	for _, section := range sections {
		if section == "" {
			continue
		}
		// Extract file path from "a/<path> b/<path>" header.
		lines := strings.SplitN(section, "\n", 2)
		if len(lines) == 0 {
			continue
		}
		header := lines[0]
		// "a/path/to/file b/path/to/file"
		parts := strings.SplitN(header, " b/", 2)
		if len(parts) != 2 {
			continue
		}
		path := strings.TrimSpace(parts[1])
		if fd, ok := fileMap[path]; ok {
			fd.Patch = "diff --git " + section
			fileMap[path] = fd
		}
	}
}

