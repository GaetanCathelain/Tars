package worker

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	wshub "github.com/GaetanCathelain/Tars/internal/ws"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	readBufSize    = 4096
	flushInterval  = 500 * time.Millisecond
	flushSizeLimit = 8 * 1024 // 8KB
)

// outputBuffer accumulates PTY output and flushes to the database in batches.
type outputBuffer struct {
	db        *pgxpool.Pool
	sessionID uuid.UUID
	mu        sync.Mutex
	buf       []byte
}

func (ob *outputBuffer) append(data []byte) {
	ob.mu.Lock()
	ob.buf = append(ob.buf, data...)
	ob.mu.Unlock()
}

func (ob *outputBuffer) shouldFlush() bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	return len(ob.buf) >= flushSizeLimit
}

func (ob *outputBuffer) flush() {
	ob.mu.Lock()
	if len(ob.buf) == 0 {
		ob.mu.Unlock()
		return
	}
	data := make([]byte, len(ob.buf))
	copy(data, ob.buf)
	ob.buf = ob.buf[:0]
	ob.mu.Unlock()

	_, err := ob.db.Exec(context.Background(),
		`INSERT INTO worker_output (session_id, data) VALUES ($1, $2)`,
		ob.sessionID, data,
	)
	if err != nil {
		slog.Error("flush worker_output", "session_id", ob.sessionID, "error", err)
	}
}

// captureOutput reads from the PTY and broadcasts output via WebSocket.
// It buffers writes to the database. Closes outputDone when finished.
func captureOutput(ctx context.Context, db *pgxpool.Pool, hub *wshub.Hub, sessionID, taskID uuid.UUID, ptmx *os.File, outputDone chan struct{}) {
	defer close(outputDone)

	ob := &outputBuffer{
		db:        db,
		sessionID: sessionID,
	}

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	// Flush on timer in background
	flushDone := make(chan struct{})
	go func() {
		defer close(flushDone)
		for {
			select {
			case <-ticker.C:
				ob.flush()
			case <-ctx.Done():
				return
			}
		}
	}()

	buf := make([]byte, readBufSize)
	for {
		n, err := ptmx.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])

			// Broadcast immediately via WebSocket (base64 encoded)
			encoded := base64.StdEncoding.EncodeToString(chunk)
			hub.BroadcastToTask(taskID, &wshub.OutgoingMessage{
				Type:      "worker_output",
				TaskID:    taskID,
				SessionID: sessionID,
				Data:      encoded,
			})

			// Buffer for DB write
			ob.append(chunk)

			// Flush if buffer is large enough
			if ob.shouldFlush() {
				ob.flush()
			}
		}
		if err != nil {
			if err != io.EOF {
				slog.Debug("pty read", "session_id", sessionID, "error", err)
			}
			break
		}
	}

	// Final flush
	ob.flush()

	// Wait for flush goroutine to stop
	<-flushDone
}
