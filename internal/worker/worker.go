// Package worker manages PTY-based Claude Code worker sessions.
// It spawns claude CLI processes in real PTYs, captures their output,
// broadcasts it via WebSocket, and persists it for replay.
package worker
