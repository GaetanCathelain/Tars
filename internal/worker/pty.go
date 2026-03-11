package worker

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// spawnClaude starts the claude CLI in a PTY with the given prompt.
// Returns the PTY master fd, the command (for Wait), and any error.
func spawnClaude(prompt string) (*os.File, *exec.Cmd, error) {
	cmd := exec.Command("claude", "--print", "--permission-mode", "bypassPermissions", prompt)
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLUMNS=120",
		"LINES=40",
	)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: 40,
		Cols: 120,
	})
	if err != nil {
		return nil, nil, err
	}

	return ptmx, cmd, nil
}
