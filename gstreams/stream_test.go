package gstreams

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"math/rand"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestBuffer(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		err := From(context.TODO(), func(ctx context.Context, source chan<- int) error {
			for i := 0; i < 10; i++ {
				if i == 5 {
					return fmt.Errorf("err1")
				}
				source <- i
			}
			return nil
		}).Buffer(2).ForEach(func(ctx context.Context, item int) {
			if item > 4 {
				t.Fatal("item>4")
			}
		})
		assert.Error(t, err)
	})
}

func TestBufferNegative(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		reduce, err := Reduce(Just(context.TODO(), 1, 2, 3, 4), func(ctx context.Context, item int) (int, error) {
			result += item
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
			elements []int
		}{
			{
				name: "no elements with nil",
			},
			{
				name:     "no elements",
				elements: []int{},
			},
			{
				name:     "1 element",
				elements: []int{1},
			},
			{
				name:     "multiple elements",
				elements: []int{1, 2, 3},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				val, err := Just(context.TODO(), test.elements...).Count()
				assert.Equal(t, len(test.elements), val)
				assert.Nil(t, err)
			})
		}
	})
}

func TestDone(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var count int32
		err := Walk(Just(context.TODO(), 1, 2, 3), func(ctx context.Context, item int, pipe chan<- int) error {
			time.Sleep(time.Millisecond * 100)
			atomic.AddInt32(&count, int32(item))
			return nil
		}).Done()
		assert.Nil(t, err)
		assert.Equal(t, int32(6), count)
	})
}

func TestDistinct(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		v, err := Reduce(Distinct(Just(context.TODO(), 4, 1, 3, 2, 3, 4), func(ctx context.Context, item int) (int, error) {
			return item, nil
		}), func(ctx context.Context, item int) (int, error) {
			result += item
			return result, nil
		})
		assert.Equal(t, 10, v)
		assert.Nil(t, err)
	})
}

func TestFilter(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		reduce, err := Reduce(Just(context.TODO(), 1, 2, 3, 4).Filter(func(ctx context.Context, item int) (bool, error) {
			return item%2 == 0, nil
		}), func(ctx context.Context, item int) (int, error) {
			result += item
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
		f, err := Just(context.TODO(), emptyArray...).First()
		assert.Nil(t, f)
		assert.Nil(t, err)
		f, err = Just(context.TODO(), "foo").First()
		assert.Equal(t, "foo", f)
		assert.Nil(t, err)
		f, err = Just(context.TODO(), "foo", "bar").First()
		assert.Equal(t, "foo", f)
		assert.Nil(t, err)
	})
}

func TestGroup(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var groups [][]int
		err := Group(Just(context.TODO(), 10, 11, 20, 21), func(ctx context.Context, item int) (int, error) {
			v := item
			return v / 10, nil
		}).ForEach(func(ctx context.Context, item []int) {
			v := item
			var group []int
			for _, each := range v {
				group = append(group, each)
			}
			groups = append(groups, group)
		})
		assert.Nil(t, err)
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
		reduce, err := Reduce(Just(context.TODO(), 1, 2, 3, 4).Head(2), func(ctx context.Context, item int) (int, error) {
			result += item
			return result, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 3, reduce)
	})
}

func TestHeadZero(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assert.Panics(t, func() {
			var result int
			_, err := Reduce(Just(context.TODO(), 1, 2, 3, 4).Head(0), func(ctx context.Context, item int) (int, error) {
				result += item
				return result, nil
			})
			assert.Nil(t, err)
		})
	})
}

func TestLast(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		goroutines := runtime.NumGoroutine()
		v, err := Just(context.TODO(), emptyArray...).Last()
		assert.Nil(t, v)
		assert.Nil(t, err)
		v, err = Just(context.TODO(), "foo").Last()
		assert.Nil(t, err)
		assert.Equal(t, "foo", v)
		v, err = Just(context.TODO(), "foo", "bar").Last()
		assert.Nil(t, err)
		assert.Equal(t, "bar", v)
		// let scheduler schedule first
		runtime.Gosched()
		assert.Equal(t, goroutines, runtime.NumGoroutine())
	})
}

func TestMerge(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		err := Merge(Just(context.TODO(), 1, 2, 3, 4)).ForEach(func(ctx context.Context, item []int) {
			assert.ElementsMatch(t, []int{1, 2, 3, 4}, item)
		})
		assert.Nil(t, err)
	})
}

func TestParallelJust(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var count int32
		err := Just(context.TODO(), 1, 2, 3).Parallel(func(ctx context.Context, item int) error {
			time.Sleep(time.Millisecond * 100)
			atomic.AddInt32(&count, int32(item))
			return nil
		}, UnlimitedWorkers())
		assert.Nil(t, err)
		assert.Equal(t, int32(6), count)
	})
}

func TestReverse(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		err := Merge(Just(context.TODO(), 1, 2, 3, 4).Reverse()).ForEach(func(ctx context.Context, item []int) {
			assert.ElementsMatch(t, []int{4, 3, 2, 1}, item)
		})
		assert.Nil(t, err)
	})
}

func TestSort(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var prev int
		err := Just(context.TODO(), 5, 3, 7, 1, 9, 6, 4, 8, 2).Sort(func(ctx context.Context, a, b int) (bool, error) {
			return a < b, nil
		}).ForEach(func(ctx context.Context, item int) {
			next := item
			assert.True(t, prev < next)
			prev = next
		})
		assert.Nil(t, err)
	})
}

func TestSplit(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assert.Panics(t, func() {
			err := Split(Just(context.TODO(), 1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 0).Done()
			assert.Nil(t, err)
		})
		var chunks [][]int
		err := Split(Just(context.TODO(), 1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 4).ForEach(func(ctx context.Context, item []int) {
			chunk := item
			chunks = append(chunks, chunk)
		})
		assert.Nil(t, err)
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
		reduce, err := Reduce(Just(context.TODO(), 1, 2, 3, 4).Tail(2), func(ctx context.Context, item int) (int, error) {
			result += item
			return result, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 7, reduce)
	})
}

func TestTailZero(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		assert.Panics(t, func() {
			_, err := Reduce(Just(context.TODO(), 1, 2, 3, 4).Tail(0), func(ctx context.Context, item int) (int, error) {
				return 0, nil
			})
			assert.Nil(t, err)
		})
	})
}

func TestWalk(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		var result int
		err := Walk(Just(context.TODO(), 1, 2, 3, 4, 5), func(ctx context.Context, item int, pipe chan<- int) error {
			if item%2 != 0 {
				pipe <- item
			}
			return nil
		}, UnlimitedWorkers()).ForEach(func(ctx context.Context, item int) {
			result += item
		})
		assert.Nil(t, err)
		assert.Equal(t, 9, result)
	})
}

func TestStream_AnyMach(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		match, err := Just(context.TODO(), 1, 2, 3).AnyMatch(func(ctx context.Context, item int) (bool, error) {
			return item == 4, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, false, match)

		match, err = Just(context.TODO(), 1, 2, 3).AnyMatch(func(ctx context.Context, item int) (bool, error) {
			return item == 0, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, false, match)

		match, err = Just(context.TODO(), 1, 2, 3).AnyMatch(func(ctx context.Context, item int) (bool, error) {
			return item == 2, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, true, match)
	})
}

func TestStream_AllMach(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		match, err := Just(context.TODO(), 1, 2, 3).AllMatch(func(ctx context.Context, item int) (bool, error) {
			return true, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, true, match)

		match, err = Just(context.TODO(), 1, 2, 3).AllMatch(func(ctx context.Context, item int) (bool, error) {
			return false, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, false, match)

		match, err = Just(context.TODO(), 1, 2, 3).AllMatch(func(ctx context.Context, item int) (bool, error) {
			return item == 1, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, false, match)
	})
}

func TestStream_NoneMatch(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		match, err := Just(context.TODO(), 1, 2, 3).NoneMatch(func(ctx context.Context, item int) (bool, error) {
			return false, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, true, match)

		match, err = Just(context.TODO(), 1, 2, 3).NoneMatch(func(ctx context.Context, item int) (bool, error) {
			return true, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, false, match)

		match, err = Just(context.TODO(), 1, 2, 3).NoneMatch(func(ctx context.Context, item int) (bool, error) {
			return item == 4, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, true, match)
	})
}

func TestStream_Skip(t *testing.T) {
	runCheckedTest(t, func(t *testing.T) {
		num, err := Just(context.TODO(), 1, 2, 3, 4).Skip(1).Count()
		assert.Nil(t, err)
		assert.Equal(t, 3, num)

		num, err = Just(context.TODO(), 1, 2, 3, 4).Skip(3).Count()
		assert.Nil(t, err)
		assert.Equal(t, 1, num)
		num, err = Just(context.TODO(), 1, 2, 3, 4).Skip(0).Count()
		assert.Nil(t, err)
		assert.Equal(t, 4, num)
		num, err = Just(context.TODO(), 1, 2, 3, 4).Skip(3).Count()
		assert.Nil(t, err)
		assert.Equal(t, 1, num)
		s := Just(context.TODO(), 1, 2, 3, 4).Skip(3)
		equal(t, s, []int{4})
		assert.Panics(t, func() {
			Just(context.TODO(), 1, 2, 3, 4).Skip(-1)
		})
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
				val, _, err := Just(context.TODO(), test.elements...).Max(func(ctx context.Context, a, b int) (bool, error) {
					return a < b, nil
				})
				assert.Nil(t, err)
				assert.Equal(t, test.max, val)
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
				val, _, err := Just(context.TODO(), test.elements...).Min(func(ctx context.Context, a, b int) (bool, error) {
					return a < b, nil
				})
				assert.Nil(t, err)
				assert.Equal(t, test.min, val)
			})
		}
	})
}

func BenchmarkMapReduce(b *testing.B) {
	b.ReportAllocs()
	mapper := func(ctx context.Context, v int) (int, error) {
		return v * v, nil
	}
	var result int
	reducer := func(ctx context.Context, item int) (int, error) {
		result += item
		return result, nil
	}
	b.ResetTimer()
	_, err := Reduce(Map(From(context.TODO(), func(ctx context.Context, input chan<- int) error {
		for i := 0; i < b.N; i++ {
			input <- rand.Int()
		}
		return nil
	}), mapper), reducer)
	assert.Nil(b, err)
}

func equal[T any](t *testing.T, stream Stream[T], data []T) {
	items := make([]T, 0)
	for item := range stream.source {
		assert.Nil(t, stream.state.error())
		items = append(items, item)
	}
	assert.Nil(t, stream.state.error())
	if !reflect.DeepEqual(items, data) {
		t.Errorf(" %v, want %v", items, data)
	}
}

func runCheckedTest(t *testing.T, fn func(t *testing.T)) {
	defer goleak.VerifyNone(t)
	fn(t)
}
