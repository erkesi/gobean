package recover

import (
	"context"
	"runtime/debug"

	"github.com/erkesi/gobean/gerrors"
	"github.com/erkesi/gobean/glogs"
)

func GoWithRecover(f func()) {
	go Recover(f)()
}

func Recover(f func()) func() {
	return func() {
		if glogs.Log != nil {
			defer func() {
				if r := recover(); r != nil {
					glogs.Log.Errorf(context.TODO(), "%v",
						gerrors.NewPanicError(r, string(debug.Stack())))
				}
			}()
		}
		f()
	}
}

func RecoverForErr(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = gerrors.NewPanicError(r, string(debug.Stack()))
			return
		}
	}()
	err = f()
	return err
}
