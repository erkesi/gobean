package main

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/gstream"
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
	dataStream := gstream.NewDataStreamOfCursor[gstream.Optional[int]](context.TODO(), func(ctx context.Context) gstream.CursorNext[gstream.Optional[int]] {
		return func(ctx context.Context) gstream.CursorNext[gstream.Optional[int]] {
			next := 1
			return func(ctx context.Context) ([]gstream.Optional[int], bool, error) {
				step := 8
				start := next
				end := next + step
				next = end
				if end > len(list) {
					end = len(list)
				}
				var subList []gstream.Optional[int]
				for _, v := range list[start:end] {
					subList = append(subList, gstream.NewOptional(v))
				}
				hasNext := next < len(list)
				if !hasNext {
					subList = append(subList, gstream.NewEmptyOptional[int]())
				}
				return subList, hasNext, nil
			}
		}(ctx)
	})
	reduce := func() func(ctx context.Context, opt gstream.Optional[int]) ([]int, error) {
		total := 0
		return func(ctx context.Context, opt gstream.Optional[int]) ([]int, error) {
			if opt.IsPresent() {
				total += opt.Get()
				return nil, nil
			}
			return []int{total}, nil
		}
	}
	sink := gstream.NewMemorySink[int]()
	dataStream.Via(gstream.NewFlatMap(reduce())).To(sink)
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

	dataStream := gstream.NewDataStreamOfCursor[gstream.Optional[int]](context.TODO(), func(ctx context.Context) gstream.CursorNext[gstream.Optional[int]] {
		return func(ctx context.Context) gstream.CursorNext[gstream.Optional[int]] {
			next := 0
			return func(ctx context.Context) ([]gstream.Optional[int], bool, error) {
				step := 8
				start := next
				end := next + step
				next = end
				if end > len(list) {
					end = len(list)
				}
				var subList []gstream.Optional[int]
				for _, v := range list[start:end] {
					subList = append(subList, gstream.NewOptional(v))
				}
				hasNext := next < len(list)
				if !hasNext {
					subList = append(subList, gstream.NewEmptyOptional[int]())
				}
				return subList, hasNext, nil
			}
		}(ctx)
	})

	sink := gstream.NewStoreSink[string](3, func(ctx context.Context, ss []string) error {
		fmt.Printf("sink output: %v(%d)\n", ss, len(ss))
		return nil
	})

	dataStream.Via(gstream.NewFlatMap(func(ctx context.Context, opt gstream.Optional[int]) ([]string, error) {
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
	dataStream := gstream.NewDataStreamOfSlice(context.TODO(), []int{1, 2, 3})
	sink := gstream.NewMemorySink[int]()
	dataStream.Via(gstream.NewFilter(func(ctx context.Context, i int) (bool, error) { return i < 2, nil })).To(sink)

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
	dataStream := gstream.NewDataStreamOfCursor[int](context.TODO(), func(ctx context.Context) gstream.CursorNext[int] {
		return func(ctx context.Context) gstream.CursorNext[int] {
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

	sink := gstream.NewMemorySink[int]()
	dataStream.Via(gstream.NewFilter(func(ctx context.Context, i int) (bool, error) { return true, nil })).To(sink)
	err := dataStream.State().Wait()
	if err != nil {
		panic(err)
	}
	for _, i := range sink.Result() {
		fmt.Println(i + 0)
	}

}
