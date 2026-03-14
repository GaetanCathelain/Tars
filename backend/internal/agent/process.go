package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/creack/pty"
)

// Process manages a single running Claude Code CLI subprocess.
type Process struct {
	AgentID string
	ptmx    *os.File // PTY master fd
	cmd     *exec.Cmd
	LogStore *LogStore
	seq      atomic.Int64
	done     chan struct{}
	cancel   context.CancelFunc
}

// SpawnConfig holds the parameters for spawning an agent process.
type SpawnConfig struct {
	AgentID      string
	WorktreePath string
	Model        string
	SystemPrompt string
	Env          []string
	OnOutput     func(line LogLine) // called for each output line
	OnExit       func(exitCode int) // called when process exits
}

// Spawn starts a new Claude Code CLI process under a PTY.
func Spawn(ctx context.Context, cfg SpawnConfig) (*Process, error) {
	ctx, cancel := context.WithCancel(ctx)

	args := []string{"--print"}
	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = cfg.WorktreePath
	cmd.Env = buildEnv(cfg.Env)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("pty start: %w", err)
	}

	p := &Process{
		AgentID:  cfg.AgentID,
		ptmx:     ptmx,
		cmd:      cmd,
		LogStore: newLogStore(),
		done:     make(chan struct{}),
		cancel:   cancel,
	}

	// Write system prompt to stdin if provided.
	if cfg.SystemPrompt != "" {
		fmt.Fprintln(ptmx, cfg.SystemPrompt)
	}

	// Start output reader goroutine.
	go p.readOutput(cfg.OnOutput)

	// Wait for process exit in background.
	go func() {
		err := cmd.Wait()
		close(p.done)
		cancel()

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		}
		if cfg.OnExit != nil {
			cfg.OnExit(exitCode)
		}
	}()

	return p, nil
}

// SendInput writes text to the agent's stdin via the PTY.
func (p *Process) SendInput(text string) error {
	_, err := fmt.Fprintln(p.ptmx, text)
	return err
}

// Stop sends SIGTERM, waits up to 5s, then sends SIGKILL.
func (p *Process) Stop() {
	if p.cmd.Process != nil {
		p.cmd.Process.Signal(os.Interrupt)
	}

	select {
	case <-p.done:
		return
	case <-time.After(5 * time.Second):
		if p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
	}
}

// Done returns a channel that is closed when the process exits.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// PID returns the process ID, or 0 if not started.
func (p *Process) PID() int {
	if p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return 0
}

// readOutput reads from the PTY master and appends lines to LogStore.
func (p *Process) readOutput(onOutput func(LogLine)) {
	defer p.ptmx.Close()

	buf := make([]byte, 4096)
	var pending []byte

	for {
		n, err := p.ptmx.Read(buf)
		if n > 0 {
			pending = append(pending, buf[:n]...)
			// Emit complete lines.
			for {
				idx := indexByte(pending, '\n')
				if idx < 0 {
					break
				}
				text := string(pending[:idx+1])
				pending = pending[idx+1:]

				seq := p.seq.Add(1)
				line := LogLine{
					Seq:    seq,
					Ts:     time.Now().UTC(),
					Stream: "stdout",
					Text:   text,
				}
				p.LogStore.Append(line)
				if onOutput != nil {
					onOutput(line)
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("agent %s: pty read: %v", p.AgentID, err)
			}
			// Flush any remaining partial line.
			if len(pending) > 0 {
				seq := p.seq.Add(1)
				line := LogLine{Seq: seq, Ts: time.Now().UTC(), Stream: "stdout", Text: string(pending)}
				p.LogStore.Append(line)
				if onOutput != nil {
					onOutput(line)
				}
			}
			return
		}
	}
}

func buildEnv(extra []string) []string {
	env := os.Environ()
	return append(env, extra...)
}

func indexByte(b []byte, c byte) int {
	for i, v := range b {
		if v == c {
			return i
		}
	}
	return -1
}
