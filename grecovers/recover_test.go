package grecovers

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/gerrors"
	"github.com/erkesi/gobean/glogs"
	"strings"
	"sync"
	"testing"
)

type Log struct {
	T testing.TB
}

func (l Log) Debugf(ctx context.Context, format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func (l Log) Errorf(ctx context.Context, format string, v ...interface{}) {
	if format != "%v" {
		l.T.Fatalf(`expected log format "%s" got "%s"`, "%v", format)
	}
	if panicErr, ok := v[0].(*gerrors.PanicError); !ok || !strings.Contains(panicErr.Error(), "123") {
		l.T.Fatalf(`expected log "%s" got "%s"`, "123", panicErr.Error())
	}
}

func TestRecover(t *testing.T) {
	glogs.Init(&Log{T: t})

	Recover(func() {
		panic(123)
	})()
}

func TestGoWithRecover(t *testing.T) {
	glogs.Init(&Log{T: t})
	var wg sync.WaitGroup
	wg.Add(1)
	GoRecover(func() {
		defer wg.Done()
		panic(123)
	})
	wg.Wait()
}

func TestRecoverForErr(t *testing.T) {
	err := RecoverOfErr(func() error {
		panic(123)
	})
	if !strings.Contains(err.Error(), "123") {
		t.Fatal(err.Error())
	}
}

func TestGoRecoverWithContext(t *testing.T) {
	glogs.Init(&Log{T: t})
	ctx := context.WithValue(context.TODO(), "k1", "v2")
	var wg sync.WaitGroup
	wg.Add(1)
	GoRecoverWithContext(ctx, func() {
		defer wg.Done()
		panic(123)
	})
	wg.Wait()
}
