package recover

import (
	"context"
	"fmt"
	"testing"

	"github.com/erkesi/gobean/glogs"
)

type Log struct {
}

func (l Log) Debugf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func (l Log) Errorf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func TestRecover(t *testing.T) {
	glogs.Init(Log{})

	Recover(func() {
		panic(123)
	})()
}

func TestGoWithRecover(t *testing.T) {
	glogs.Init(Log{})

	GoWithRecover(func() {
		panic(123)
	})

	// time.Sleep(5 * time.Second)
}
