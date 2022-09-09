package recover

import (
	"context"
	"runtime/debug"

	"github.com/erkesi/gobean/glogs"
)

func GoWithRecover(f func()) {
	go Recover(f)()
}

func Recover(f func()) func() {
	return func() {
		if glogs.Log != nil {
			defer func() {
				if err := recover(); err != nil {
					glogs.Log.Debugf(context.TODO(),
						"recover: err is %v, %s", err, string(debug.Stack()))
				}
			}()
		}
		f()
	}
}
