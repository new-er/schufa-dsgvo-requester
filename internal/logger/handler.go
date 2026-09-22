package logger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"log/slog"
)

const maxLogLines = 1000
const logFileName = "schufa.log"

// RingBuffer holds the last maxLogLines log lines (newest first when snapped).
type RingBuffer struct {
	mu   sync.Mutex
	buf  [maxLogLines]string
	head int
	full bool
}

func (r *RingBuffer) Push(line string) {
	r.mu.Lock()
	r.buf[r.head] = line
	r.head = (r.head + 1) % maxLogLines
	if r.head == 0 {
		r.full = true
	}
	r.mu.Unlock()
}

// Snapshot returns a slice of log lines, newest first.
func (r *RingBuffer) Snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := maxLogLines
	if !r.full {
		n = r.head
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		idx := (r.head - 1 - i + maxLogLines) % maxLogLines
		out = append(out, r.buf[idx])
	}
	return out
}

// LogHandler implements slog.Handler writing JSON lines to a file and ring buffer.
type LogHandler struct {
	file *os.File
	rb   *RingBuffer
}

func NewLogHandler(logDir string) (*LogHandler, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	logPath := filepath.Join(logDir, logFileName)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	rb := &RingBuffer{}
	// Load existing lines (last maxLogLines) into ring buffer
	if data, err := os.ReadFile(logPath); err == nil {
		lines := splitLines(string(data))
		// keep only last maxLogLines
		start := 0
		if len(lines) > maxLogLines {
			start = len(lines) - maxLogLines
		}
		for _, line := range lines[start:] {
			rb.Push(line)
		}
	}
	return &LogHandler{file: f, rb: rb}, nil
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range splitBytes([]byte(s)) {
		lines = append(lines, string(line))
	}
	return lines
}

func splitBytes(b []byte) [][]byte {
	var res [][]byte
	start := 0
	for i, ch := range b {
		if ch == '\n' {
			res = append(res, b[start:i])
			start = i + 1
		}
	}
	if start < len(b) {
		res = append(res, b[start:])
	}
	return res
}

func (h *LogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *LogHandler) Handle(_ context.Context, r slog.Record) error {
	// Build JSON map
	m := map[string]any{
		"time":  r.Time.Format(time.RFC3339Nano),
		"level": r.Level.String(),
		"msg":   r.Message,
	}
	// Collect attributes
	r.Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.Any()
		return true
	})
	data, _ := json.Marshal(m)
	line := string(data)

	// Write to file
	if _, err := h.file.WriteString(line + "\n"); err != nil {
		// ignore write errors
	}
	h.file.Sync()
	// Push to ring buffer
	h.rb.Push(line)
	return nil
}

func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h // simple, ignore
}
func (h *LogHandler) WithGroup(name string) slog.Handler {
	return h
}

// GetLogs returns newest-first log lines for UI.
func (h *LogHandler) GetLogs() []string {
	return h.rb.Snapshot()
}