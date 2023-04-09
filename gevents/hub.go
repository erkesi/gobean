package gevents

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/erkesi/gobean/ginjects"
)

type registerOptions struct {
	priority int
}

type RegisterOption func(opt *registerOptions)

// WithRegisterPriority
// priority 越大越先执行
func WithRegisterPriority(priority int) RegisterOption {
	return func(opt *registerOptions) {
		opt.priority = priority
	}
}

func Register(executor Executor, opts ...RegisterOption) {
	opt := &registerOptions{}
	for _, f := range opts {
		f(opt)
	}
	hub.register(&executorExt{
		executor: executor,
		priority: opt.priority})
}

func SetDefaultExecutor(executor Executor) {
	ginjects.ProvideByValue(executor, ginjects.WithProvidePriorityTop1())
	hub.defaultExecutor = executor
}

func Clear() {
	hub.clear()
}

func execute(ctx context.Context, event interface{}, o *pubOptions) error {
	eventType := reflect.TypeOf(event)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}
	exts, defaultExecute := hub.findExecutes(eventType)
	if len(exts) == 0 && o.mustHaveSubscriber {
		return fmt.Errorf("gevents: event type `%T`, not find executor", event)
	}
	if len(exts) == 0 {
		if defaultExecute == nil {
			return fmt.Errorf("gevents: event type `%T`, not find executor", event)
		} else {
			return defaultExecute.Execute(ctx, event)
		}
	}
	var err error
	for _, ext := range exts {
		err = ext.executor.Execute(ctx, event)
		if err != nil {
			break
		}
	}
	return err
}

var hub = &_hub{}

type _hub struct {
	executes        map[reflect.Type][]*executorExt
	defaultExecutor Executor
}

func (h *_hub) register(ext *executorExt) {
	if h.executes == nil {
		h.executes = map[reflect.Type][]*executorExt{}
	}
	ginjects.ProvideByValue(ext.executor, ginjects.WithProvidePriorityTop1())
	eventTypeSet := map[reflect.Type]bool{}
	for _, eventType := range ext.executor.Types() {
		if eventType.Kind() == reflect.Ptr {
			eventType = eventType.Elem()
		}
		if !eventTypeSet[eventType] {
			eventTypeSet[eventType] = true
			h.executes[eventType] = append(h.executes[eventType], ext)
		}
	}
	for eventType := range eventTypeSet {
		sort.Slice(h.executes[eventType], sortExecutes(h.executes[eventType]))
	}
}

func sortExecutes(exts []*executorExt) func(i int, j int) bool {
	return func(i, j int) bool {
		return exts[i].priority > exts[j].priority
	}
}

func (h *_hub) clear() {
	h.executes = nil
	h.defaultExecutor = nil
}

func (h *_hub) findExecutes(eventType reflect.Type) ([]*executorExt, Executor) {
	return h.executes[eventType], h.defaultExecutor
}
