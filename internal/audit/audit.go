package audit

import (
	"sync"
	"time"
)

// Event 审计事件。
type Event struct {
	Time   time.Time
	Action string
	Key    string
	Detail string
}

// Log 环形审计缓冲。
type Log struct {
	mu   sync.Mutex
	cap  int
	buf  []Event
	head int
}

func New(capacity int) *Log {
	if capacity < 8 {
		capacity = 64
	}
	return &Log{cap: capacity, buf: make([]Event, 0, capacity)}
}

func (l *Log) Record(action, key, detail string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{Time: time.Now().UTC(), Action: action, Key: key, Detail: detail}
	if len(l.buf) < l.cap {
		l.buf = append(l.buf, e)
		return
	}
	l.buf[l.head] = e
	l.head = (l.head + 1) % l.cap
}

// Recent 返回最近 n 条（拷贝）。
func (l *Log) Recent(n int) []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n <= 0 || len(l.buf) == 0 {
		return nil
	}
	if n > len(l.buf) {
		n = len(l.buf)
	}
	out := make([]Event, n)
	if len(l.buf) < l.cap {
		copy(out, l.buf[len(l.buf)-n:])
		return out
	}
	// 环形：从 head 起的旧→新
	start := (l.head - n + l.cap) % l.cap
	for i := 0; i < n; i++ {
		out[i] = l.buf[(start+i)%l.cap]
	}
	return out
}
