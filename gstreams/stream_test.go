package gstreams

import (
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestBuffer(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		const N = 5
		var count int32
		var wait sync.WaitGroup
		wait.Add(1)
		From(func(source chan<- int) {
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for i := 0; i < 2*N; i++ {
				select {
				case source <- i:
					atomic.AddInt32(&count, 1)
				case <-ticker.C:
					wait.Done()
					return
				}
			}
		}).Buffer(N).ForAll(func(pipe <-chan int) {
			wait.Wait()
			// why N+1, because take one more to wait for sending into the channel
			assert.Equal(t, int32(N+1), atomic.LoadInt32(&count))
		})
	})
}

func TestBufferNegative(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		reduce, err := Reduce(Just(1, 2, 3, 4).Buffer(-1), func(pipe <-chan int) (any, error) {
			for item := range pipe {
				result += item
			}
			return result, nil
		})
		if err != nil {
			return
		}
		assert.Equal(t, 10, reduce)
	})
}

func TestCount(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		tests := []struct {
			name     string
			elements []any
		}{
			{
				name: "no elements with nil",
			},
			{
				name:     "no elements",
				elements: []any{},
			},
			{
				name:     "1 element",
				elements: []any{1},
			},
			{
				name:     "multiple elements",
				elements: []any{1, 2, 3},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				val := Just(test.elements...).Count()
				assert.Equal(t, len(test.elements), val)
			})
		}
	})
}

func TestDone(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var count int32
		Walk(Just(1, 2, 3), func(item int, pipe chan<- any) {
			time.Sleep(time.Millisecond * 100)
			atomic.AddInt32(&count, int32(item))
		}).Done()
		assert.Equal(t, int32(6), count)
	})
}

func TestJust(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		reduce, err := Reduce(Just(1, 2, 3, 4), func(pipe <-chan int) (any, error) {
			for item := range pipe {
				result += item
			}
			return result, nil
		})
		if err != nil {
			return
		}
		assert.Equal(t, 10, reduce)
	})
}

func TestDistinct(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		Reduce(Distinct(Just(4, 1, 3, 2, 3, 4), func(item int) int {
			return item
		}), func(pipe <-chan int) (any, error) {
			for item := range pipe {
				result += item
			}
			return result, nil
		})
		assert.Equal(t, 10, result)
	})
}

func TestFilter(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		reduce, err := Reduce(Just(1, 2, 3, 4).Filter(func(item int) bool {
			return item%2 == 0
		}), func(pipe <-chan int) (int, error) {
			for item := range pipe {
				result += item
			}
			return result, nil
		})
		if err != nil {
			return
		}
		assert.Equal(t, 6, reduce)
	})
}

var emptyArray []interface{}

func TestFirst(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assert.Nil(t, Just(emptyArray...).First())
		assert.Equal(t, "foo", Just("foo").First())
		assert.Equal(t, "foo", Just("foo", "bar").First())
	})
}

func TestForAll(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		Just(1, 2, 3, 4).Filter(func(item int) bool {
			return item%2 == 0
		}).ForAll(func(pipe <-chan int) {
			for item := range pipe {
				result += item
			}
		})
		assert.Equal(t, 6, result)
	})
}

func TestGroup(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var groups [][]int
		Group(Just(10, 11, 20, 21), func(item int) int {
			v := item
			return v / 10
		}).ForEach(func(item []int) {
			v := item
			var group []int
			for _, each := range v {
				group = append(group, each)
			}
			groups = append(groups, group)
		})

		assert.Equal(t, 2, len(groups))
		for _, group := range groups {
			assert.Equal(t, 2, len(group))
			assert.True(t, group[0]/10 == group[1]/10)
		}
	})
}

func TestHead(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		Reduce(Just(1, 2, 3, 4).Head(2), func(pipe <-chan int) (int, error) {
			for item := range pipe {
				result += item
			}
			return result, nil
		})
		assert.Equal(t, 3, result)
	})
}

func TestHeadZero(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assert.Panics(t, func() {
			Reduce(Just(1, 2, 3, 4).Head(0), func(pipe <-chan int) (int, error) {
				return 0, nil
			})
		})
	})
}

func TestHeadMore(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		Reduce(Just(1, 2, 3, 4).Head(6), func(pipe <-chan int) (int, error) {
			for item := range pipe {
				result += item
			}
			return result, nil
		})
		assert.Equal(t, 10, result)
	})
}

func TestLast(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		goroutines := runtime.NumGoroutine()
		assert.Nil(t, Just(emptyArray...).Last())
		assert.Equal(t, "foo", Just("foo").Last())
		assert.Equal(t, "bar", Just("foo", "bar").Last())
		// let scheduler schedule first
		runtime.Gosched()
		assert.Equal(t, goroutines, runtime.NumGoroutine())
	})
}

func TestMerge(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		Merge(Just(1, 2, 3, 4)).ForEach(func(item []int) {
			assert.ElementsMatch(t, []any{1, 2, 3, 4}, item)
		})
	})
}

func TestParallelJust(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var count int32
		Just(1, 2, 3).Parallel(func(item int) {
			time.Sleep(time.Millisecond * 100)
			atomic.AddInt32(&count, int32(item))
		}, UnlimitedWorkers())
		assert.Equal(t, int32(6), count)
	})
}

func TestReverse(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		Merge(Just(1, 2, 3, 4).Reverse()).ForEach(func(item []int) {
			assert.ElementsMatch(t, []any{4, 3, 2, 1}, item)
		})
	})
}

func TestSort(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var prev int
		Just(5, 3, 7, 1, 9, 6, 4, 8, 2).Sort(func(a, b int) bool {
			return a < b
		}).ForEach(func(item int) {
			next := item
			assert.True(t, prev < next)
			prev = next
		})
	})
}

func TestSplit(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assert.Panics(t, func() {
			Split(Just(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 0).Done()
		})
		var chunks [][]int
		Split(Just(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 4).ForEach(func(item []int) {
			chunk := item
			chunks = append(chunks, chunk)
		})
		assert.EqualValues(t, [][]int{
			{1, 2, 3, 4},
			{5, 6, 7, 8},
			{9, 10},
		}, chunks)
	})
}

func TestTail(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		Reduce(Just(1, 2, 3, 4).Tail(2), func(pipe <-chan int) (int, error) {
			for item := range pipe {
				result += item
			}
			return result, nil
		})
		assert.Equal(t, 7, result)
	})
}

func TestTailZero(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assert.Panics(t, func() {
			Reduce(Just(1, 2, 3, 4).Tail(0), func(pipe <-chan int) (int, error) {
				return 0, nil
			})
		})
	})
}

func TestWalk(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		Walk(Just(1, 2, 3, 4, 5), func(item int, pipe chan<- int) {
			if item%2 != 0 {
				pipe <- item
			}
		}, UnlimitedWorkers()).ForEach(func(item int) {
			result += item
		})
		assert.Equal(t, 9, result)
	})
}

func TestStream_AnyMach(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assetEqual(t, false, Just(1, 2, 3).AnyMatch(func(item int) bool {
			return item == 4
		}))
		assetEqual(t, false, Just(1, 2, 3).AnyMatch(func(item int) bool {
			return item == 0
		}))
		assetEqual(t, true, Just(1, 2, 3).AnyMatch(func(item int) bool {
			return item == 2
		}))
		assetEqual(t, true, Just(1, 2, 3).AnyMatch(func(item int) bool {
			return item == 2
		}))
	})
}

func TestStream_AllMach(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assetEqual(
			t, true, Just(1, 2, 3).AllMatch(func(item int) bool {
				return true
			}),
		)
		assetEqual(
			t, false, Just(1, 2, 3).AllMatch(func(item int) bool {
				return false
			}),
		)
		assetEqual(
			t, false, Just(1, 2, 3).AllMatch(func(item int) bool {
				return item == 1
			}),
		)
	})
}

func TestStream_NoneMatch(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assetEqual(
			t, true, Just(1, 2, 3).NoneMatch(func(item int) bool {
				return false
			}),
		)
		assetEqual(
			t, false, Just(1, 2, 3).NoneMatch(func(item int) bool {
				return true
			}),
		)
		assetEqual(
			t, true, Just(1, 2, 3).NoneMatch(func(item int) bool {
				return item == 4
			}),
		)
	})
}

func TestConcat(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		a1 := []any{1, 2, 3}
		a2 := []any{4, 5, 6}
		s1 := Just(a1...)
		s2 := Just(a2...)
		stream := Concat(s1, s2)
		var items []any
		for item := range stream.source {
			items = append(items, item)
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].(int) < items[j].(int)
		})
		ints := make([]any, 0)
		ints = append(ints, a1...)
		ints = append(ints, a2...)
		assetEqual(t, ints, items)
	})
}

func TestStream_Skip(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assetEqual(t, 3, Just(1, 2, 3, 4).Skip(1).Count())
		assetEqual(t, 1, Just(1, 2, 3, 4).Skip(3).Count())
		assetEqual(t, 4, Just(1, 2, 3, 4).Skip(0).Count())
		equal(t, Just(1, 2, 3, 4).Skip(3), []int{4})
		assert.Panics(t, func() {
			Just(1, 2, 3, 4).Skip(-1)
		})
	})
}

func TestStream_Concat(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		stream := Just(1).Concat(Just(2), Just(3))
		var items []int
		for item := range stream.source {
			items = append(items, item)
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i] < items[j]
		})
		assetEqual(t, []int{1, 2, 3}, items)

		just := Just(1)
		equal(t, just.Concat(just), []int{1})
	})
}

func TestStream_Max(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		tests := []struct {
			name     string
			elements []int
			max      int
		}{
			{
				name: "no elements with nil",
			},
			{
				name:     "no elements",
				elements: []int{},
				max:      0,
			},
			{
				name:     "1 element",
				elements: []int{1},
				max:      1,
			},
			{
				name:     "multiple elements",
				elements: []int{1, 2, 9, 5, 8},
				max:      9,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				val, _ := Just(test.elements...).Max(func(a, b int) bool {
					return a < b
				})
				assetEqual(t, test.max, val)
			})
		}
	})
}

func TestStream_Min(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		tests := []struct {
			name     string
			elements []int
			min      int
		}{
			{
				name: "no elements with nil",
				min:  0,
			},
			{
				name:     "no elements",
				elements: []int{},
				min:      0,
			},
			{
				name:     "1 element",
				elements: []int{1},
				min:      1,
			},
			{
				name:     "multiple elements",
				elements: []int{-1, 1, 2, 9, 5, 8},
				min:      -1,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				val, _ := Just(test.elements...).Min(func(a, b int) bool {
					return a < b
				})
				assetEqual(t, test.min, val)
			})
		}
	})
}

func BenchmarkParallelMapReduce(b *testing.B) {
	b.ReportAllocs()

	mapper := func(v any) any {
		return v.(int64) * v.(int64)
	}
	reducer := func(input <-chan any) (any, error) {
		var result int64
		for v := range input {
			result += v.(int64)
		}
		return result, nil
	}
	b.ResetTimer()
	Reduce(Map(From(func(input chan<- any) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				input <- int64(rand.Int())
			}
		})
	}), mapper), reducer)
}

func BenchmarkMapReduce(b *testing.B) {
	b.ReportAllocs()

	mapper := func(v any) any {
		return v.(int64) * v.(int64)
	}
	reducer := func(input <-chan any) (any, error) {
		var result int64
		for v := range input {
			result += v.(int64)
		}
		return result, nil
	}
	b.ResetTimer()
	Reduce(Map(From(func(input chan<- any) {
		for i := 0; i < b.N; i++ {
			input <- int64(rand.Int())
		}
	}), mapper), reducer)
}

func assetEqual(t *testing.T, except, data any) {
	if !reflect.DeepEqual(except, data) {
		t.Errorf(" %v, want %v", data, except)
	}
}

func equal[T any](t *testing.T, stream Stream[T], data []T) {
	items := make([]T, 0)
	for item := range stream.source {
		items = append(items, item)
	}
	if !reflect.DeepEqual(items, data) {
		t.Errorf(" %v, want %v", items, data)
	}
}

func runCheckedTest(t *testing.T, fn func(t *testing.T)) {
	defer goleak.VerifyNone(t)
	fn(t)
}
