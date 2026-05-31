package runner

import (
	"context"
	"sync"
)

// Pool is a small bounded-parallelism helper. Submit functions; Wait
// blocks until all submitted work completes. Errors are collected and
// returned together — one failing job does not cancel others.
type Pool struct {
	sem chan struct{}
	wg  sync.WaitGroup

	mu   sync.Mutex
	errs []error
}

// NewPool returns a pool with at most max concurrent jobs. If max <= 0
// it falls back to 1.
func NewPool(max int) *Pool {
	if max <= 0 {
		max = 1
	}
	return &Pool{sem: make(chan struct{}, max)}
}

// Go submits fn. Returns immediately. If ctx is cancelled before a
// slot is available, fn is not run and ctx.Err() is recorded.
func (p *Pool) Go(ctx context.Context, fn func(context.Context) error) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		select {
		case p.sem <- struct{}{}:
		case <-ctx.Done():
			p.recordErr(ctx.Err())
			return
		}
		defer func() { <-p.sem }()
		if err := fn(ctx); err != nil {
			p.recordErr(err)
		}
	}()
}

func (p *Pool) recordErr(err error) {
	p.mu.Lock()
	p.errs = append(p.errs, err)
	p.mu.Unlock()
}

// Wait blocks until all submitted jobs complete and returns collected errors.
func (p *Pool) Wait() []error {
	p.wg.Wait()
	return p.errs
}
