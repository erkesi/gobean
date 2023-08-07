package gstreams

import (
	"context"
	"github.com/erkesi/gobean/grecovers"
	"sort"
	"sync"
)

const (
	minWorkers = 1
)

var empty emptyType

type emptyType = struct{}

type (
	options struct {
		unlimitedWorkers bool
		workers          int
	}
	// FilterFunc defines the method to filter a Stream.
	FilterFunc[T any] func(ctx context.Context, item T) (bool, error)
	// GenerateFunc defines the method to send elements into a Stream.
	GenerateFunc[T any] func(ctx context.Context, source chan<- T) error
	// ForEachFunc defines the method to handle each element in a Stream.
	ForEachFunc[T any] func(ctx context.Context, item T)
	// KeyFunc defines the method to generate keys for the elements in a Stream.
	KeyFunc[T, K any] func(ctx context.Context, item T) (K, error)
	// MapFunc defines the method to map each element to another object in a Stream.
	MapFunc[T, R any] func(ctx context.Context, item T) (R, error)
	// FlatMapFunc defines the method to map each element to another object in a Stream.
	FlatMapFunc[T, R any] func(ctx context.Context, item T) ([]R, error)
	// LessFunc defines the method to compare the elements in a Stream.
	LessFunc[T any] func(ctx context.Context, a, b T) (bool, error)
	// PredicateFunc defines the method to predicate a Stream.
	PredicateFunc[T any] func(ctx context.Context, item T) (bool, error)
	// Option defines the method to customize a Stream.
	Option func(opts *options)
	// ParallelFunc defines the method to handle elements parallel.
	ParallelFunc[T any] func(ctx context.Context, item T) error
	// ReduceFunc defines the method to reduce all the elements in a Stream.
	ReduceFunc[T, R any] func(ctx context.Context, item T) (R, error)
	// WalkFunc defines the method to walk through all the elements in a Stream.
	WalkFunc[T, R any] func(ctx context.Context, item T, pipe chan<- R) error

	// A Stream is a Stream that can be used to do Stream processing.
	Stream[T any] struct {
		ctx    context.Context
		source <-chan T
		state  *streamState
	}

	streamState struct {
		m   sync.RWMutex
		err error
	}
)

func (s *streamState) error() error {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.err
}

func (s *streamState) setError(err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.err == nil {
		s.err = err
	}
}

// From constructs a Stream[T] from the given GenerateFunc.
func From[T any](ctx context.Context, generate GenerateFunc[T]) Stream[T] {
	source := make(chan T)
	state := &streamState{}
	go func() {
		defer close(source)
		err := grecovers.RecoverFn(func() error {
			return generate(ctx, source)
		})()
		if err != nil {
			state.setError(err)
			return
		}
	}()
	return _range(ctx, source, state)
}

// Just converts the given arbitrary items to a Stream.
func Just[T any](ctx context.Context, items ...T) Stream[T] {
	source := make(chan T, len(items))
	for _, item := range items {
		source <- item
	}
	close(source)
	return _range(ctx, source, &streamState{})
}

// Filter filters the items by the given FilterFunc.
func (s Stream[T]) Filter(fn FilterFunc[T], opts ...Option) Stream[T] {
	return Walk(s, func(ctx context.Context, item T, pipe chan<- T) error {
		b, err := fn(ctx, item)
		if err != nil {
			return err
		}
		if b {
			pipe <- item
		}
		return nil
	}, opts...)
}

// Map converts each item to another corresponding item, which means it's a 1:1 model.
func Map[T, R any](s Stream[T], fn MapFunc[T, R], opts ...Option) Stream[R] {
	return Walk(s, func(ctx context.Context, item T, pipe chan<- R) error {
		v, err := fn(ctx, item)
		if err != nil {
			return err
		}
		pipe <- v
		return nil
	}, opts...)
}

// FlatMap converts each item to another corresponding item, which means it's a 1:N model.
func FlatMap[T, R any](s Stream[T], fn FlatMapFunc[T, R], opts ...Option) Stream[R] {
	return Walk(s, func(ctx context.Context, item T, pipe chan<- R) error {
		v, err := fn(ctx, item)
		if err != nil {
			return err
		}
		for _, o := range v {
			pipe <- o
		}
		return nil
	}, opts...)
}

// Reduce is a utility method to let the caller deal with the underlying channel.
func Reduce[T, R any](s Stream[T], fn ReduceFunc[T, R]) (R, error) {
	var r R
	for item := range s.source {
		var r0 R
		if err := s.state.error(); err != nil {
			cleanCh(s.source)
			return r0, err
		}
		r0, err := fn(s.ctx, item)
		if err != nil {
			cleanCh(s.source)
			return r0, err
		}
		r = r0
	}
	return r, s.state.error()
}

// Walk lets the callers handle each item, the caller may write zero, one or more items base on the given item.
func Walk[T, R any](s Stream[T], fn WalkFunc[T, R], opts ...Option) Stream[R] {
	option := buildOptions(opts...)
	if option.unlimitedWorkers {
		return walkUnlimited(s, fn, option)
	}
	return walkLimited(s, fn, option)
}

// Distinct removes the duplicated items base on the given KeyFunc.
func Distinct[T any, K comparable](s Stream[T], fn KeyFunc[T, K]) Stream[T] {
	source := make(chan T)
	go func() {
		defer close(source)
		keys := make(map[K]struct{})
		for item := range s.source {
			if err := s.state.error(); err != nil {
				go cleanCh(s.source)
				return
			}
			key, err := grecovers.RecoverVGFn(func() (interface{}, error) {
				return fn(s.ctx, item)
			})()
			if err != nil {
				s.state.setError(err)
				go cleanCh(s.source)
				return
			}
			if _, ok := keys[key.(K)]; !ok {
				source <- item
				keys[key.(K)] = struct{}{}
			}
		}
	}()
	return _range(s.ctx, source, s.state)
}

// Group groups the elements into different groups based on their keys.
func Group[T any, K comparable](s Stream[T], fn KeyFunc[T, K]) Stream[[]T] {
	groups := make(map[K][]T)
	for item := range s.source {
		if err := s.state.error(); err != nil {
			go cleanCh(s.source)
			break
		}
		key, err := grecovers.RecoverVGFn(func() (interface{}, error) {
			return fn(s.ctx, item)
		})()
		if err != nil {
			s.state.setError(err)
			go cleanCh(s.source)
			break
		}
		groups[key.(K)] = append(groups[key.(K)], item)
	}
	source := make(chan []T)
	go func() {
		defer close(source)
		for _, group := range groups {
			source <- group
		}
	}()
	return _range(s.ctx, source, s.state)
}

// Chunk splits the elements into chunk with size up to n,
// might be less than n on tailing elements.
func Chunk[T any](s Stream[T], n int) Stream[[]T] {
	if n < 1 {
		panic("n should be greater than 0")
	}
	source := make(chan []T)
	go func() {
		defer close(source)
		var chunk []T
		for item := range s.source {
			if s.state.error() != nil {
				go cleanCh(s.source)
				return
			}
			chunk = append(chunk, item)
			if len(chunk) == n {
				source <- chunk
				chunk = nil
			}
		}
		if chunk != nil {
			source <- chunk
		}
	}()
	return _range(s.ctx, source, s.state)
}

// Merge merges all the items into a slice and generates a new Stream.
func Merge[T any](s Stream[T]) Stream[[]T] {
	var items []T
	for item := range s.source {
		if s.state.error() != nil {
			go cleanCh(s.source)
			break
		}
		items = append(items, item)
	}
	source := make(chan []T, 1)
	source <- items
	close(source)
	return _range(s.ctx, source, s.state)
}

// AllMatch returns whether all elements of this Stream[T] match the provided predicate.
// May not evaluate the predicate on all elements if not necessary for determining the result.
// If the Stream[T] is empty then true is returned and the predicate is not evaluated.
func (s Stream[T]) AllMatch(predicate PredicateFunc[T]) (bool, error) {
	for item := range s.source {
		if err := s.state.error(); err != nil {
			go cleanCh(s.source)
			return false, err
		}
		b, err := predicate(s.ctx, item)
		if !b || err != nil {
			go cleanCh(s.source)
			return false, err
		}
	}
	if err := s.state.error(); err != nil {
		return false, err
	}
	return true, nil
}

// AnyMatch returns whether any elements of this Stream[T] match the provided predicate.
// May not evaluate the predicate on all elements if not necessary for determining the result.
// If the Stream[T] is empty then false is returned and the predicate is not evaluated.
func (s Stream[T]) AnyMatch(predicate PredicateFunc[T]) (bool, error) {
	for item := range s.source {
		if err := s.state.error(); err != nil {
			go cleanCh(s.source)
			return false, err
		}
		b, err := predicate(s.ctx, item)
		if err != nil {
			go cleanCh(s.source)
			return false, err
		}
		if b {
			go cleanCh(s.source)
			return true, nil
		}
	}
	return false, s.state.error()
}

// NoneMatch returns whether all elements of this Stream[T] don't match the provided predicate.
// May not evaluate the predicate on all elements if not necessary for determining the result.
// If the Stream[T] is empty then true is returned and the predicate is not evaluated.
func (s Stream[T]) NoneMatch(predicate PredicateFunc[T]) (bool, error) {
	for item := range s.source {
		if err := s.state.error(); err != nil {
			return false, err
		}
		b, err := predicate(s.ctx, item)
		if err != nil {
			go cleanCh(s.source)
			return false, err
		}
		if b {
			go cleanCh(s.source)
			return false, nil
		}
	}
	if err := s.state.error(); err != nil {
		return false, err
	}
	return true, nil
}

// Buffer buffers the items into a queue with size n.
// It can balance the producer and the consumer if their processing throughput don't match.
func (s Stream[T]) Buffer(n int) Stream[T] {
	if n < 0 {
		n = 0
	}
	source := make(chan T, n)
	go func() {
		defer close(source)
		for item := range s.source {
			source <- item
		}
	}()
	return _range(s.ctx, source, s.state)
}

// Parallel applies the given ParallelFunc to each item concurrently with given number of workers.
func (s Stream[T]) Parallel(fn ParallelFunc[T], opts ...Option) error {
	return Walk(s, func(ctx context.Context, item T, pipe chan<- T) error {
		return fn(ctx, item)
	}, opts...).Done()
}

// Count counts the number of elements in the result.
func (s Stream[T]) Count() (int, error) {
	count := 0
	for range s.source {
		if err := s.state.error(); err != nil {
			go cleanCh(s.source)
			return 0, s.state.error()
		}
		count++
	}
	return count, s.state.error()
}

// Done waits all upStream[T]ing operations to be done.
func (s Stream[T]) Done() error {
	cleanCh(s.source)
	return s.state.error()
}

// First returns the first item, nil if no items.
func (s Stream[T]) First() (T, error) {
	var t T
	for item := range s.source {
		go cleanCh(s.source)
		return item, s.state.error()
	}
	return t, s.state.error()
}

// ForEach seals the Stream[T] with the ForEachFunc on each item, no successive operations.
func (s Stream[T]) ForEach(fn ForEachFunc[T]) error {
	for item := range s.source {
		if err := s.state.error(); err != nil {
			go cleanCh(s.source)
			return err
		}
		fn(s.ctx, item)
	}
	return s.state.error()
}

// Collect seals the Stream[T] with the collect on each item.
func (s Stream[T]) Collect() ([]T, error) {
	var ts []T
	for item := range s.source {
		if err := s.state.error(); err != nil {
			go cleanCh(s.source)
			return ts, err
		}
		ts = append(ts, item)
	}
	return ts, s.state.error()
}

// CollectToMap map the elements into different value based on their key.
func CollectToMap[T any, K comparable](s Stream[T], fn KeyFunc[T, K]) (map[K]T, error) {
	m := make(map[K]T)
	for item := range s.source {
		if err := s.state.error(); err != nil {
			go cleanCh(s.source)
			return m, err
		}
		if k, err := fn(s.ctx, item); err != nil {
			go cleanCh(s.source)
			return m, err
		} else {
			m[k] = item
		}
	}
	return m, s.state.error()
}

// Head returns the first n elements in p.
func (s Stream[T]) Head(n int) Stream[T] {
	if n < 1 {
		panic("n must be greater than 0")
	}
	source := make(chan T)
	go func() {
		defer close(source)
		i, num := 0, n
		for item := range s.source {
			i++
			if err := s.state.error(); err != nil {
				go cleanCh(s.source)
				return
			}
			num--
			if num >= 0 {
				source <- item
			}
			if num == 0 {
				go cleanCh(s.source)
				return
			}
		}
	}()
	return _range(s.ctx, source, s.state)
}

// Last returns the last item, or nil if no items.
func (s Stream[T]) Last() (T, error) {
	var item T
	for item = range s.source {
	}
	return item, s.state.error()
}

// Reverse reverses the elements in the Stream.
func (s Stream[T]) Reverse() Stream[T] {
	var items []T
	for item := range s.source {
		if s.state.error() != nil {
			go cleanCh(s.source)
			break
		}
		items = append(items, item)
	}
	// reverse, official method
	for i := len(items)/2 - 1; i >= 0; i-- {
		opp := len(items) - 1 - i
		items[i], items[opp] = items[opp], items[i]
	}
	source := make(chan T)
	go func() {
		defer close(source)
		for _, item := range items {
			source <- item
		}
	}()
	return _range(s.ctx, source, s.state)
}

// Max returns the maximum item from the underlying source.
func (s Stream[T]) Max(less LessFunc[T]) (T, bool, error) {
	var max T
	var i int
	for item := range s.source {
		if s.state.error() != nil {
			go cleanCh(s.source)
			break
		}
		i++
		if i == 1 {
			max = item
			continue
		}
		b, err := less(s.ctx, max, item)
		if err != nil {
			go cleanCh(s.source)
			return max, false, err
		}
		if b {
			max = item
		}
	}
	return max, i > 0, s.state.error()
}

// Min returns the minimum item from the underlying source.
func (s Stream[T]) Min(less LessFunc[T]) (T, bool, error) {
	var min T
	var i int
	for item := range s.source {
		if s.state.error() != nil {
			go cleanCh(s.source)
			break
		}
		i++
		if i == 1 {
			min = item
			continue
		}
		b, err := less(s.ctx, item, min)
		if err != nil {
			go cleanCh(s.source)
			return min, false, err
		}
		if b {
			min = item
		}
	}
	return min, i > 0, s.state.error()
}

// Sort sorts the items from the underlying source.
func (s Stream[T]) Sort(less LessFunc[T]) Stream[T] {
	var items []T
	for item := range s.source {
		if s.state.error() != nil {
			go cleanCh(s.source)
			break
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		b, err := less(s.ctx, items[i], items[j])
		if err != nil {
			s.state.setError(err)
			return b
		}
		return b
	})
	source := make(chan T)
	go func() {
		defer close(source)
		for _, item := range items {
			source <- item
		}
	}()
	return _range(s.ctx, source, s.state)
}

// Skip returns a Stream[T] that skips size elements.
func (s Stream[T]) Skip(n int64) Stream[T] {
	if n < 0 {
		panic("n must not be negative")
	}
	if n == 0 {
		return s
	}
	source := make(chan T)
	go func() {
		defer close(source)
		for item := range s.source {
			if s.state.error() != nil {
				go cleanCh(s.source)
				return
			}
			n--
			if n >= 0 {
				continue
			} else {
				source <- item
			}
		}
	}()

	return _range(s.ctx, source, s.state)
}

// Tail returns the last n elements in p.
func (s Stream[T]) Tail(n int) Stream[T] {
	if n < 1 {
		panic("n should be greater than 0")
	}

	source := make(chan T)

	go func() {
		defer close(source)
		var array []T
		for item := range s.source {
			if s.state.error() != nil {
				go cleanCh(s.source)
				return
			}
			array = append(array, item)
		}
		for _, item := range array[len(array)-n:] {
			source <- item
		}
	}()

	return _range(s.ctx, source, s.state)
}

func walkLimited[T, R any](s Stream[T], fn WalkFunc[T, R], option *options) Stream[R] {
	pipe := make(chan R, option.workers)
	go func() {
		defer close(pipe)
		var wg sync.WaitGroup
		sem := make(chan emptyType, option.workers)
		for item := range s.source {
			if s.state.error() != nil {
				go cleanCh(s.source)
				return
			}
			val := item
			sem <- empty
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					<-sem
				}()
				err := grecovers.RecoverFn(func() error {
					return fn(s.ctx, val, pipe)
				})()
				if err != nil {
					s.state.setError(err)
					return
				}
			}()
		}
		wg.Wait()
	}()
	return _range(s.ctx, pipe, s.state)
}

func walkUnlimited[T, R any](s Stream[T], fn WalkFunc[T, R], option *options) Stream[R] {
	pipe := make(chan R, option.workers)
	go func() {
		defer close(pipe)
		var wg sync.WaitGroup
		for item := range s.source {
			if s.state.error() != nil {
				go cleanCh(s.source)
				return
			}
			val := item
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := grecovers.RecoverFn(func() error {
					return fn(s.ctx, val, pipe)
				})()
				if err != nil {
					s.state.setError(err)
					return
				}
			}()
		}
		wg.Wait()
	}()
	return _range(s.ctx, pipe, s.state)
}

// _range converts the given channel to a Stream.
func _range[T any](ctx context.Context, source <-chan T, state *streamState) Stream[T] {
	return Stream[T]{
		ctx:    ctx,
		source: source,
		state:  state,
	}
}

// UnlimitedWorkers lets the caller use as many workers as the tasks.
func UnlimitedWorkers() Option {
	return func(opts *options) {
		opts.unlimitedWorkers = true
	}
}

// WithWorkers lets the caller customize the concurrent workers.
func WithWorkers(workers int) Option {
	return func(opts *options) {
		if workers < minWorkers {
			opts.workers = minWorkers
		} else {
			opts.workers = workers
		}
	}
}

// buildOptions returns an options with given customizations.
func buildOptions(opts ...Option) *options {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	return options
}

// cleanCh the given channel.
func cleanCh[T any](channel <-chan T) {
	for range channel {
	}
}

// newOptions returns a default options.
func newOptions() *options {
	return &options{
		workers: minWorkers,
	}
}
