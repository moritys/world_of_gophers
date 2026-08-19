package limiter

import "sync/atomic"

type Limiter struct {
	counter atomic.Int64
	limit   int64
}

func NewLimiter(limit int64) *Limiter {
	return &Limiter{limit: limit}
}

func (l *Limiter) Acquire() bool {
	for {
		cur := l.counter.Load()
		if cur >= l.limit {
			return false
		}
		if l.counter.CompareAndSwap(cur, cur+1) {
			return true
		}
	}
}

func (l *Limiter) Release() {
	l.counter.Add(-1)
}

func (l *Limiter) InUse() int64 {
	return l.counter.Load()
}
