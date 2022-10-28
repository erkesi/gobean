package gthreads

import (
	"context"
	"sync"
)

type ValueGroup struct {
	cancel  func()
	wg      sync.WaitGroup
	mu      sync.Mutex
	errOnce sync.Once
	err     error
	res     []interface{}
}

// WithContext returns a new Group and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func WithContext(ctx context.Context) (*ValueGroup, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &ValueGroup{cancel: cancel}, ctx
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *ValueGroup) Wait() ([]interface{}, error) {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.res, g.err
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *ValueGroup) Go(f func() (interface{}, error)) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if val, err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		} else {
			g.mu.Lock()
			defer g.mu.Unlock()
			g.res = append(g.res, val)
		}
	}()
}
