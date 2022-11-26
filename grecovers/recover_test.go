package grecovers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/erkesi/gobean/gerrors"
	"github.com/erkesi/gobean/glogs"
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

	GoRecover(func() {
		panic(123)
	})
	time.Sleep(2 * time.Second)
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
	GoRecoverWithContext(ctx, func() {
		panic(123)
	})
	time.Sleep(2 * time.Second)
}
