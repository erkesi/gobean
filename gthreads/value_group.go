package gthreads

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
)

type ValueGroup struct {
	cancel  func()
	wg      sync.WaitGroup
	mu      sync.Mutex
	errOnce sync.Once
	err     error
	vals    []*orderVal
	order   int64
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
	if g.err != nil {
		return nil, g.err
	}
	sort.Slice(g.vals, func(i, j int) bool {
		return g.vals[i].order < g.vals[j].order
	})
	res := make([]interface{}, 0, len(g.vals))
	for _, val := range g.vals {
		res = append(res, val.val)
	}
	return res, nil
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *ValueGroup) Go(f func() (interface{}, error)) {
	order := atomic.AddInt64(&g.order, 1)
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
			g.vals = append(g.vals, &orderVal{
				order: order,
				val:   val,
			})
		}
	}()
}

type orderVal struct {
	order int64
	val   interface{}
}
