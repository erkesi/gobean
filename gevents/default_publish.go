package gevents

import (
	"context"
	"fmt"
	"runtime/debug"
)

type DefaultPublisher struct {
}

func (ep *DefaultPublisher) Publish(ctx context.Context, event interface{}, opts ...PublishOpt) error {
	var err error
	defer func() {
		r := recover()
		if r != nil {
			msg := string(debug.Stack())
			err = fmt.Errorf("%v\n%s", r, msg)
			return
		}
	}()
	o := &Option{}
	for _, opt := range opts {
		opt(o)
	}
	err = execute(ctx, event, o)
	return err
}
