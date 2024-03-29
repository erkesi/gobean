package grecovers

import (
	"context"
	"runtime/debug"

	"github.com/erkesi/gobean/gerrors"
	"github.com/erkesi/gobean/glogs"
)

func GoRecover(f func()) {
	go Recover(f)()
}

func GoRecoverWithContext(ctx context.Context, f func()) {
	go RecoverWithContext(ctx, f)()
}

func RecoverWithContext(ctx context.Context, f func()) func() {
	return func() {
		if glogs.Log != nil {
			defer func() {
				if r := recover(); r != nil {
					glogs.Log.Errorf(ctx, "%v",
						gerrors.NewPanicError(r, string(debug.Stack())))
				}
			}()
		}
		f()
	}
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

func RecoverOfErr(f func() error) (err error) {
	return RecoverFn(f)()
}

func RecoverFn(fn func() error) func() error {
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = gerrors.NewPanicError(r, string(debug.Stack()))
				return
			}
		}()
		err = fn()
		return
	}
}

func RecoverVGFn(fn func() (interface{}, error)) func() (interface{}, error) {
	return func() (val interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = gerrors.NewPanicError(r, string(debug.Stack()))
				return
			}
		}()
		val, err = fn()
		return
	}
}
