package main

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/gstream"
)

func main() {
	//testNewDataStreamOf()
	testNewDataStreamOfCursor()
}

func testNewDataStreamOf() {
	dataStream := gstream.NewDataStreamOf(context.TODO(), []int{1, 2, 3})
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
