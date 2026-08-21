package limit

import (
	"sync"
	"time"
)

// TokenBucket 简易限流（管理页试写保护）。
type TokenBucket struct {
	mu     sync.Mutex
	rate   float64
	burst  float64
	tokens float64
	last   time.Time
}

func NewBucket(rate float64, burst float64) *TokenBucket {
	if rate <= 0 {
		rate = 10
	}
	if burst <= 0 {
		burst = rate
	}
	return &TokenBucket{rate: rate, burst: burst, tokens: burst, last: time.Now()}
}

func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.last = now
	b.tokens += elapsed * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Concurrency 信号量。
type Concurrency struct {
	ch chan struct{}
}

func NewConcurrency(n int) *Concurrency {
	if n < 1 {
		n = 1
	}
	return &Concurrency{ch: make(chan struct{}, n)}
}

func (c *Concurrency) Acquire() { c.ch <- struct{}{} }
func (c *Concurrency) Release() { <-c.ch }
