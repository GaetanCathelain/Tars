package agent

import (
	"sync"
	"time"
)

const logCapacity = 10000

// LogLine is a single line of agent output.
type LogLine struct {
	Seq    int64     `json:"seq"`
	Ts     time.Time `json:"ts"`
	Stream string    `json:"stream"` // "stdout" or "stderr"
	Text   string    `json:"text"`
}

// LogStore is a thread-safe ring buffer of agent output lines.
type LogStore struct {
	mu    sync.RWMutex
	lines []LogLine
	total int64 // monotonic count of all lines ever appended
}

func newLogStore() *LogStore {
	return &LogStore{
		lines: make([]LogLine, 0, logCapacity),
	}
}

// Append adds a line to the ring buffer.
func (s *LogStore) Append(line LogLine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.lines) >= logCapacity {
		// Overwrite oldest entry (ring behaviour).
		copy(s.lines, s.lines[1:])
		s.lines[logCapacity-1] = line
	} else {
		s.lines = append(s.lines, line)
	}
	s.total++
}

// Get returns up to limit lines starting at offset, and the total count ever appended.
func (s *LogStore) Get(limit, offset int) ([]LogLine, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.total
	n := len(s.lines)

	if offset >= n {
		return []LogLine{}, total
	}

	end := offset + limit
	if end > n {
		end = n
	}

	result := make([]LogLine, end-offset)
	copy(result, s.lines[offset:end])
	return result, total
}
