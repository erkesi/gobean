package main

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/gstreamings"
	. "github.com/erkesi/gobean/gstreams"
)

func testStreams() {
	reduceFunc := func(ctx context.Context, i int64) (int64, error) {
		var t int64
		t += i
		return t, nil
	}
	r, err := Reduce(Map(Just(context.Background(), 1, 2, 3), func(ctx context.Context, item int) (int64, error) {
		return int64(item), nil
	}).Skip(1).Tail(2), reduceFunc)
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
}

func main() {
	testNewDataStreamOf()
	testNewDataStreamOfCursor()
	testOptional()
	testStoreSink()
	testStreams()
}

func testOptional() {
	var list []int
	for i := 0; i < 10; i++ {
		list = append(list, i)
	}
	reduce := func() func(ctx context.Context, opt gstreamings.Optional[int]) ([]int, error) {
		total := 0
		return func(ctx context.Context, opt gstreamings.Optional[int]) ([]int, error) {
			if opt.IsPresent() {
				total += opt.Get()
				return nil, nil
			}
			return []int{total}, nil
		}
	}
	sink := gstreamings.NewMemorySink[int]()
	err := gstreamings.NewStreamingOfCursor[gstreamings.Optional[int]](context.TODO(), func(ctx context.Context) gstreamings.CursorNext[gstreamings.Optional[int]] {
		return func(ctx context.Context) gstreamings.CursorNext[gstreamings.Optional[int]] {
			next := 1
			return func(ctx context.Context) ([]gstreamings.Optional[int], bool, error) {
				step := 8
				start := next
				end := next + step
				next = end
				if end > len(list) {
					end = len(list)
				}
				var subList []gstreamings.Optional[int]
				for _, v := range list[start:end] {
					subList = append(subList, gstreamings.NewOptional(v))
				}
				hasNext := next < len(list)
				if !hasNext {
					subList = append(subList, gstreamings.NewEmptyOptional[int]())
				}
				return subList, hasNext, nil
			}
		}(ctx)
	}).Via(gstreamings.NewFlatMap(reduce())).To(sink).Wait()
	if err != nil {
		panic(err)
	}
	for _, i := range sink.Result() {
		fmt.Println(i + 0)
	}
}

func testStoreSink() {
	var list []int
	for i := 0; i < 10; i++ {
		list = append(list, i)
	}
	sink := gstreamings.NewStoreSink[string](3, func(ctx context.Context, ss []string) error {
		fmt.Printf("sink output: %v(%d)\n", ss, len(ss))
		return nil
	})
	err := gstreamings.NewStreamingOfCursor[gstreamings.Optional[int]](context.TODO(), func(ctx context.Context) gstreamings.CursorNext[gstreamings.Optional[int]] {
		return func(ctx context.Context) gstreamings.CursorNext[gstreamings.Optional[int]] {
			next := 0
			return func(ctx context.Context) ([]gstreamings.Optional[int], bool, error) {
				step := 8
				start := next
				end := next + step
				next = end
				if end > len(list) {
					end = len(list)
				}
				var subList []gstreamings.Optional[int]
				for _, v := range list[start:end] {
					subList = append(subList, gstreamings.NewOptional(v))
				}
				hasNext := next < len(list)
				if !hasNext {
					subList = append(subList, gstreamings.NewEmptyOptional[int]())
				}
				return subList, hasNext, nil
			}
		}(ctx)
	}).Via(gstreamings.NewFlatMap(func(ctx context.Context, opt gstreamings.Optional[int]) ([]string, error) {
		if opt.IsPresent() {
			return []string{fmt.Sprint(opt.Get())}, nil
		}
		return nil, nil
	})).To(sink).Wait()
	if err != nil {
		panic(err)
	}
}

func testNewDataStreamOf() {
	sink := gstreamings.NewMemorySink[int]()
	err := gstreamings.NewStreamingOfSlice(context.TODO(), []int{1, 2, 3}).
		Via(gstreamings.NewFilter(func(ctx context.Context, i int) (bool, error) { return i < 2, nil })).
		To(sink).Wait()
	if err != nil {
		panic(err)
	}
	for _, i := range sink.Result() {
		fmt.Println(i + 0)
	}
}

func testNewDataStreamOfCursor() {
	var list []int
	for i := 0; i < 100; i++ {
		list = append(list, i*10)
	}
	sink := gstreamings.NewMemorySink[int]()
	err := gstreamings.NewStreamingOfCursor[int](context.TODO(), func(ctx context.Context) gstreamings.CursorNext[int] {
		return func(ctx context.Context) gstreamings.CursorNext[int] {
			next := 1
			return func(ctx context.Context) ([]int, bool, error) {
				step := 8
				start := next
				end := next + step
				next = end
				if end > len(list) {
					end = len(list)
				}
				subList := list[start:end]
				return subList, next < len(list), nil
			}
		}(ctx)
	}).Via(gstreamings.NewFilter(func(ctx context.Context, i int) (bool, error) { return true, nil })).
		To(sink).Wait()
	if err != nil {
		panic(err)
	}
	for _, i := range sink.Result() {
		fmt.Println(i + 0)
	}

}
