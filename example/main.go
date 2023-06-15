package main

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/gstreamings"
)

func main() {
	//testNewDataStreamOf()
	//testNewDataStreamOfCursor()
	//testOptional()
	testStoreSink()
}

func testOptional() {
	var list []int
	for i := 0; i < 10; i++ {
		list = append(list, i)
	}
	dataStream := gstreamings.NewDataStreamOfCursor[gstreamings.Optional[int]](context.TODO(), func(ctx context.Context) gstreamings.CursorNext[gstreamings.Optional[int]] {
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
	})
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
	dataStream.Via(gstreamings.NewFlatMap(reduce())).To(sink)
	err := dataStream.State().Wait()
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

	dataStream := gstreamings.NewDataStreamOfCursor[gstreamings.Optional[int]](context.TODO(), func(ctx context.Context) gstreamings.CursorNext[gstreamings.Optional[int]] {
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
	})

	sink := gstreamings.NewStoreSink[string](3, func(ctx context.Context, ss []string) error {
		fmt.Printf("sink output: %v(%d)\n", ss, len(ss))
		return nil
	})

	dataStream.Via(gstreamings.NewFlatMap(func(ctx context.Context, opt gstreamings.Optional[int]) ([]string, error) {
		if opt.IsPresent() {
			return []string{fmt.Sprint(opt.Get())}, nil
		}
		return nil, nil
	})).To(sink)

	err := dataStream.State().Wait()
	if err != nil {
		panic(err)
	}
}

func testNewDataStreamOf() {
	dataStream := gstreamings.NewDataStreamOfSlice(context.TODO(), []int{1, 2, 3})
	sink := gstreamings.NewMemorySink[int]()
	dataStream.Via(gstreamings.NewFilter(func(ctx context.Context, i int) (bool, error) { return i < 2, nil })).To(sink)

	err := dataStream.State().Wait()
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
	dataStream := gstreamings.NewDataStreamOfCursor[int](context.TODO(), func(ctx context.Context) gstreamings.CursorNext[int] {
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
	})

	sink := gstreamings.NewMemorySink[int]()
	dataStream.Via(gstreamings.NewFilter(func(ctx context.Context, i int) (bool, error) { return true, nil })).To(sink)
	err := dataStream.State().Wait()
	if err != nil {
		panic(err)
	}
	for _, i := range sink.Result() {
		fmt.Println(i + 0)
	}

}
