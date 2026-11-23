package web

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// SSEMessage represents a Server-Sent Event message with optional event name and data payload.
type SSEMessage struct {
	Event string
	Data  string
}

// LogEntry stores a single timestamped log or terminal output line.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Text      string    `json:"text"`
	Type      string    `json:"type"` // "info", "warn", "error", "command", "process"
}

// Broadcaster manages active SSE client channels and maintains a recent log history buffer.
type Broadcaster struct {
	mu          sync.RWMutex
	clients     map[chan SSEMessage]bool
	history     []LogEntry
	historySize int
}

// NewBroadcaster creates a new Broadcaster instance with the specified history buffer capacity.
func NewBroadcaster(historySize int) *Broadcaster {
	if historySize <= 0 {
		historySize = 250
	}
	return &Broadcaster{
		clients:     make(map[chan SSEMessage]bool),
		history:     make([]LogEntry, 0, historySize),
		historySize: historySize,
	}
}

// Subscribe registers a new SSE listener channel.
func (b *Broadcaster) Subscribe() chan SSEMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan SSEMessage, 100)
	b.clients[ch] = true
	return ch
}

// Unsubscribe removes an SSE listener channel and closes it.
func (b *Broadcaster) Unsubscribe(ch chan SSEMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
}

// Broadcast sends an SSE message to all connected clients.
func (b *Broadcaster) Broadcast(event, data string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msg := SSEMessage{Event: event, Data: data}
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
			// Buffer full, drop to prevent blocking
		}
	}
}

// AppendLog adds a log entry to history and broadcasts it as an SSE event.
func (b *Broadcaster) AppendLog(entryType, text string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Text:      text,
		Type:      entryType,
	}

	b.mu.Lock()
	if len(b.history) >= b.historySize {
		b.history = append(b.history[1:], entry)
	} else {
		b.history = append(b.history, entry)
	}
	b.mu.Unlock()

	formattedHTML := fmt.Sprintf(
		`<div class="log-line log-%s font-mono text-xs leading-relaxed break-all"><span class="opacity-50">[%s]</span> %s</div>`,
		entryType,
		entry.Timestamp.Format("15:04:05"),
		escapeHTML(text),
	)

	b.Broadcast("log", formattedHTML)
}

// GetHistory returns a copy of all current log entries.
func (b *Broadcaster) GetHistory() []LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	res := make([]LogEntry, len(b.history))
	copy(res, b.history)
	return res
}

// LogWriter returns an io.Writer that forwards written lines to AppendLog.
func (b *Broadcaster) LogWriter(entryType string) io.Writer {
	return &broadcasterWriter{b: b, entryType: entryType}
}

type broadcasterWriter struct {
	b         *Broadcaster
	entryType string
	buf       bytes.Buffer
	mu        sync.Mutex
}

func (w *broadcasterWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n = len(p)
	w.buf.Write(p)

	for {
		line, err := w.buf.ReadString('\n')
		if err != nil {
			// Incomplete line, put back into buffer
			w.buf.WriteString(line)
			break
		}
		trimmed := trimTrailingCRLF(line)
		if trimmed != "" {
			w.b.AppendLog(w.entryType, trimmed)
		}
	}
	return n, nil
}

func trimTrailingCRLF(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\r' || s[len(s)-1] == '\n') {
		s = s[:len(s)-1]
	}
	return s
}

func escapeHTML(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
