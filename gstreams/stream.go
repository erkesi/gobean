package gstreams

import (
	"sort"
	"sync"
)

const (
	defaultWorkers = 16
	minWorkers     = 1
)

var empty emptyType

type emptyType = struct{}

type (
	options struct {
		unlimitedWorkers bool
		workers          int
	}

	// FilterFunc defines the method to filter a Stream.
	FilterFunc[T any] func(item T) bool
	// ForAllFunc defines the method to handle all elements in a Stream.
	ForAllFunc[T any] func(pipe <-chan T)
	// ForEachFunc defines the method to handle each element in a Stream.
	ForEachFunc[T any] func(item T)
	// GenerateFunc defines the method to send elements into a Stream.
	GenerateFunc[T any] func(source chan<- T)
	// KeyFunc defines the method to generate keys for the elements in a Stream.
	KeyFunc[T, R any] func(item T) R
	// LessFunc defines the method to compare the elements in a Stream.
	LessFunc[T any] func(a, b T) bool
	// MapFunc defines the method to map each element to another object in a Stream.
	MapFunc[T, R any] func(item T) R
	// FlatMapFunc defines the method to map each element to another object in a Stream.
	FlatMapFunc[T, R any] func(item T) []R
	// Option defines the method to customize a Stream.
	Option func(opts *options)
	// ParallelFunc defines the method to handle elements parallel.
	ParallelFunc[T any] func(item T)
	// ReduceFunc defines the method to reduce all the elements in a Stream.
	ReduceFunc[T, R any] func(pipe <-chan T) (R, error)
	// WalkFunc defines the method to walk through all the elements in a Stream.
	WalkFunc[T, R any] func(item T, pipe chan<- R)

	// A Stream is a Stream that can be used to do Stream processing.
	Stream[T any] struct {
		source <-chan T
	}
)

// From constructs a Stream[T] from the given GenerateFunc.
func From[T any](generate GenerateFunc[T]) Stream[T] {
	source := make(chan T)
	go func() {
		defer close(source)
		generate(source)
	}()
	return Range(source)
}

// Just converts the given arbitrary items to a Stream.
func Just[T any](items ...T) Stream[T] {
	source := make(chan T, len(items))
	for _, item := range items {
		source <- item
	}
	close(source)
	return Range(source)
}

// Range converts the given channel to a Stream.
func Range[T any](source <-chan T) Stream[T] {
	return Stream[T]{
		source: source,
	}
}

// Concat returns a concatenated Stream.
func Concat[T any](s Stream[T], others ...Stream[T]) Stream[T] {
	return s.Concat(others...)
}

// Map converts each item to another corresponding item, which means it's a 1:1 model.
func Map[T, R any](s Stream[T], fn MapFunc[T, R], opts ...Option) Stream[R] {
	return Walk(s, func(item T, pipe chan<- R) {
		pipe <- fn(item)
	}, opts...)
}

// FlatMap converts each item to another corresponding item, which means it's a 1:N model.
func FlatMap[T, R any](s Stream[T], fn FlatMapFunc[T, R], opts ...Option) Stream[R] {
	return Walk(s, func(item T, pipe chan<- R) {
		for _, v := range fn(item) {
			pipe <- v
		}
	}, opts...)
}

// Reduce is a utility method to let the caller deal with the underlying channel.
func Reduce[T, R any](s Stream[T], fn ReduceFunc[T, R]) (R, error) {
	return fn(s.source)
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
			key := fn(item)
			if _, ok := keys[key]; !ok {
				source <- item
				keys[key] = struct{}{}
			}
		}
	}()

	return Range(source)
}

// Group groups the elements into different groups based on their keys.
func Group[T any, K comparable](s Stream[T], fn KeyFunc[T, K]) Stream[[]T] {
	groups := make(map[K][]T)
	for item := range s.source {
		key := fn(item)
		groups[key] = append(groups[key], item)
	}

	source := make(chan []T)
	go func() {
		for _, group := range groups {
			source <- group
		}
		close(source)
	}()

	return Range(source)
}

// Split splits the elements into chunk with size up to n,
// might be less than n on tailing elements.
func Split[T any](s Stream[T], n int) Stream[[]T] {
	if n < 1 {
		panic("n should be greater than 0")
	}

	source := make(chan []T)
	go func() {
		var chunk []T
		for item := range s.source {
			chunk = append(chunk, item)
			if len(chunk) == n {
				source <- chunk
				chunk = nil
			}
		}
		if chunk != nil {
			source <- chunk
		}
		close(source)
	}()

	return Range(source)
}

// Merge merges all the items into a slice and generates a new Stream.
func Merge[T any](s Stream[T]) Stream[[]T] {
	var items []T
	for item := range s.source {
		items = append(items, item)
	}

	source := make(chan []T, 1)
	source <- items
	close(source)

	return Range(source)
}

// AllMatch returns whether all elements of this Stream[T] match the provided predicate.
// May not evaluate the predicate on all elements if not necessary for determining the result.
// If the Stream[T] is empty then true is returned and the predicate is not evaluated.
func (s Stream[T]) AllMatch(predicate func(item T) bool) bool {
	for item := range s.source {
		if !predicate(item) {
			// make sure the former goroutine not block, and current func returns fast.
			go drain(s.source)
			return false
		}
	}
	return true
}

// AnyMatch returns whether any elements of this Stream[T] match the provided predicate.
// May not evaluate the predicate on all elements if not necessary for determining the result.
// If the Stream[T] is empty then false is returned and the predicate is not evaluated.
func (s Stream[T]) AnyMatch(predicate func(item T) bool) bool {
	for item := range s.source {
		if predicate(item) {
			// make sure the former goroutine not block, and current func returns fast.
			go drain(s.source)
			return true
		}
	}

	return false
}

// Buffer buffers the items into a queue with size n.
// It can balance the producer and the consumer if their processing throughput don't match.
func (s Stream[T]) Buffer(n int) Stream[T] {
	if n < 0 {
		n = 0
	}

	source := make(chan T, n)
	go func() {
		for item := range s.source {
			source <- item
		}
		close(source)
	}()

	return Range(source)
}

// Concat returns a Stream[T] that concatenated other Stream[T]s
func (s Stream[T]) Concat(others ...Stream[T]) Stream[T] {
	source := make(chan T)

	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range s.source {
				source <- item
			}
		}()

		for _, each := range others {
			each := each
			wg.Add(1)
			go func() {
				defer wg.Done()
				for item := range each.source {
					source <- item
				}
			}()
		}

		wg.Wait()
		close(source)
	}()

	return Range(source)
}

// Parallel applies the given ParallelFunc to each item concurrently with given number of workers.
func (s Stream[T]) Parallel(fn ParallelFunc[T], opts ...Option) {
	Walk(s, func(item T, pipe chan<- T) {
		fn(item)
	}, opts...).Done()
}

// Count counts the number of elements in the result.
func (s Stream[T]) Count() (count int) {
	for range s.source {
		count++
	}
	return
}

// Done waits all upStream[T]ing operations to be done.
func (s Stream[T]) Done() {
	drain(s.source)
}

// Filter filters the items by the given FilterFunc.
func (s Stream[T]) Filter(fn FilterFunc[T], opts ...Option) Stream[T] {
	return Walk(s, func(item T, pipe chan<- T) {
		if fn(item) {
			pipe <- item
		}
	}, opts...)
}

// First returns the first item, nil if no items.
func (s Stream[T]) First() T {
	var t T
	for item := range s.source {
		// make sure the former goroutine not block, and current func returns fast.
		go drain(s.source)
		return item
	}
	return t
}

// ForAll handles the Stream[T]ing elements from the source and no later Stream[T]s.
func (s Stream[T]) ForAll(fn ForAllFunc[T]) {
	fn(s.source)
	// avoid goroutine leak on fn not consuming all items.
	go drain(s.source)
}

// ForEach seals the Stream[T] with the ForEachFunc on each item, no successive operations.
func (s Stream[T]) ForEach(fn ForEachFunc[T]) {
	for item := range s.source {
		fn(item)
	}
}

// Head returns the first n elements in p.
func (s Stream[T]) Head(n int64) Stream[T] {
	if n < 1 {
		panic("n must be greater than 0")
	}

	source := make(chan T)

	go func() {
		for item := range s.source {
			n--
			if n >= 0 {
				source <- item
			}
			if n == 0 {
				// let successive method go ASAP even we have more items to skip
				close(source)
				// why we don't just break the loop, and drain to consume all items.
				// because if breaks, this former goroutine will block forever,
				// which will cause goroutine leak.
				drain(s.source)
			}
		}
		// not enough items in s.source, but we need to let successive method to go ASAP.
		if n > 0 {
			close(source)
		}
	}()

	return Range(source)
}

// Last returns the last item, or nil if no items.
func (s Stream[T]) Last() (item T) {
	for item = range s.source {
	}
	return
}

// Max returns the maximum item from the underlying source.
func (s Stream[T]) Max(less LessFunc[T]) (T, bool) {
	var max T
	var i int
	for item := range s.source {
		i++
		if i == 1 {
			max = item
			continue
		}
		if less(max, item) {
			max = item
		}
	}

	return max, i > 0
}

// Min returns the minimum item from the underlying source.
func (s Stream[T]) Min(less LessFunc[T]) (T, bool) {
	var min T
	var i int
	for item := range s.source {
		i++
		if i == 1 {
			min = item
			continue
		}
		if less(item, min) {
			min = item
		}
	}

	return min, i > 0
}

// NoneMatch returns whether all elements of this Stream[T] don't match the provided predicate.
// May not evaluate the predicate on all elements if not necessary for determining the result.
// If the Stream[T] is empty then true is returned and the predicate is not evaluated.
func (s Stream[T]) NoneMatch(predicate func(item T) bool) bool {
	for item := range s.source {
		if predicate(item) {
			// make sure the former goroutine not block, and current func returns fast.
			go drain(s.source)
			return false
		}
	}

	return true
}

// Reverse reverses the elements in the Stream.
func (s Stream[T]) Reverse() Stream[T] {
	var items []T
	for item := range s.source {
		items = append(items, item)
	}
	// reverse, official method
	for i := len(items)/2 - 1; i >= 0; i-- {
		opp := len(items) - 1 - i
		items[i], items[opp] = items[opp], items[i]
	}

	return Just(items...)
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
		for item := range s.source {
			n--
			if n >= 0 {
				continue
			} else {
				source <- item
			}
		}
		close(source)
	}()

	return Range(source)
}

// Sort sorts the items from the underlying source.
func (s Stream[T]) Sort(less LessFunc[T]) Stream[T] {
	var items []T
	for item := range s.source {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return less(items[i], items[j])
	})

	return Just(items...)
}

// Tail returns the last n elements in p.
func (s Stream[T]) Tail(n int) Stream[T] {
	if n < 1 {
		panic("n should be greater than 0")
	}

	source := make(chan T)

	go func() {
		var array []T
		for item := range s.source {
			array = append(array, item)
		}
		for _, item := range array[len(array)-n:] {
			source <- item
		}
		close(source)
	}()

	return Range(source)
}

func walkLimited[T, R any](s Stream[T], fn WalkFunc[T, R], option *options) Stream[R] {
	pipe := make(chan R, option.workers)

	go func() {
		var wg sync.WaitGroup
		sem := make(chan emptyType, option.workers)

		for item := range s.source {
			// important, used in another goroutine
			val := item
			sem <- empty
			wg.Add(1)

			go func() {
				defer func() {
					wg.Done()
					<-sem
				}()

				fn(val, pipe)
			}()
		}

		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)
}

func walkUnlimited[T, R any](s Stream[T], fn WalkFunc[T, R], option *options) Stream[R] {
	pipe := make(chan R, option.workers)

	go func() {
		var wg sync.WaitGroup

		for item := range s.source {
			// important, used in another goroutine
			val := item
			wg.Add(1)
			go func() {
				defer wg.Done()
				fn(val, pipe)
			}()
		}

		wg.Wait()
		close(pipe)
	}()

	return Range(pipe)
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

// drain drains the given channel.
func drain[T any](channel <-chan T) {
	for range channel {
	}
}

// newOptions returns a default options.
func newOptions() *options {
	return &options{
		workers: defaultWorkers,
	}
}
