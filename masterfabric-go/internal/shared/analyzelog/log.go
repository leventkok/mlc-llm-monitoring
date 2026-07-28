package analyzelog

import (
	"sync"
	"time"
)

const defaultCapacity = 100

// Entry is one analyze attempt for the admin log monitor.
type Entry struct {
	Time      time.Time `json:"time"`
	UserID    string    `json:"user_id"`
	ReviewID  string    `json:"review_id"`
	Category  string    `json:"category,omitempty"`
	Sentiment string    `json:"sentiment,omitempty"`
	LatencyMs int       `json:"latency_ms,omitempty"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
}

type ring struct {
	mu       sync.RWMutex
	capacity int
	items    []Entry
}

var defaultLog = newRing(defaultCapacity)

func newRing(capacity int) *ring {
	if capacity < 1 {
		capacity = defaultCapacity
	}
	return &ring{capacity: capacity, items: make([]Entry, 0, capacity)}
}

func (r *ring) add(e Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.items) >= r.capacity {
		r.items = r.items[1:]
	}
	r.items = append(r.items, e)
}

func (r *ring) recent(limit int) []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if limit <= 0 || limit > len(r.items) {
		limit = len(r.items)
	}
	start := len(r.items) - limit
	out := make([]Entry, limit)
	copy(out, r.items[start:])
	return out
}

// Record appends an analyze log entry.
func Record(e Entry) {
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}
	defaultLog.add(e)
}

// Recent returns the latest entries (newest last).
func Recent(limit int) []Entry {
	return defaultLog.recent(limit)
}
