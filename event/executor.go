package event

import (
	"context"
	"reflect"
)

type Executor interface {
	Execute(ctx context.Context, event interface{}) error
	Types() []reflect.Type
}
