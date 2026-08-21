package clock

import (
	"sync"
	"time"

	"example.com/objmeta/internal/meta"
)

type system struct{}

func (system) Now() time.Time { return time.Now().UTC() }

// System 返回墙钟。
func System() meta.Clock { return system{} }

// Fake 可注入测试时钟。
type Fake struct {
	mu   sync.Mutex
	now  time.Time
	step time.Duration
}

func NewFake(start time.Time) *Fake {
	return &Fake{now: start.UTC(), step: time.Millisecond}
}

func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := f.now
	f.now = f.now.Add(f.step)
	return t
}

func (f *Fake) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = t.UTC()
}

func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}
