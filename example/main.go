package main

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/gstream"
)

func main() {
	dataStream := gstream.NewDataStreamOf(context.TODO(), []int{1, 2, 3})

	sink := gstream.NewMemorySink[int]()
	dataStream.Via(gstream.NewFilter(func(ctx context.Context, i int) bool { return i < 2 }, 1)).To(sink)

	err := dataStream.State().Wait()
	if err != nil {
		panic(err)
	}
	for _, i := range sink.Result() {
		fmt.Println(i + 0)
	}
}
