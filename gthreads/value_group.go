package gthreads

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

type token struct{}

// A ValueGroup is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero ValueGroup is valid, has no limit on the number of active goroutines,
// and does not cancel on error.
type ValueGroup struct {
	cancel func()

	wg sync.WaitGroup

	sem chan token

	errOnce sync.Once
	err     error

	mu    sync.Mutex
	vals  []*orderVal
	order int64
}

func (g *ValueGroup) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
}

// WithContext returns a new ValueGroup and an associated Context derived from ctx.
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
// It blocks until the new goroutine can be added without the number of
// active goroutines in the group exceeding the configured limit.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *ValueGroup) Go(f func() (interface{}, error)) {
	if g.sem != nil {
		g.sem <- token{}
	}
	order := atomic.AddInt64(&g.order, 1)
	g.wg.Add(1)
	go func() {
		defer g.done()

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

// TryGo calls the given function in a new goroutine only if the number of
// active goroutines in the group is currently below the configured limit.
//
// The return value reports whether the goroutine was started.
func (g *ValueGroup) TryGo(f func() (interface{}, error)) bool {
	if g.sem != nil {
		select {
		case g.sem <- token{}:
			// Note: this allows barging iff channels in general allow barging.
		default:
			return false
		}
	}
	order := atomic.AddInt64(&g.order, 1)
	g.wg.Add(1)
	go func() {
		defer g.done()
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
	return true
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
func (g *ValueGroup) SetLimit(n int) {
	if n < 0 {
		g.sem = nil
		return
	}
	if len(g.sem) != 0 {
		panic(fmt.Errorf("gthreads: modify limit while %v goroutines in the value group are still active", len(g.sem)))
	}
	g.sem = make(chan token, n)
}

type orderVal struct {
	order int64
	val   interface{}
}
