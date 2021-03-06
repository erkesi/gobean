package gevents

import (
	"context"
	"fmt"
	"github.com/erkesi/gobean/ginjects"
	"reflect"
)

func Register(executors ...Executor) {
	hub.register(executors...)
}

func SetDefaultExecutor(executor Executor) {
	ginjects.ProvideByValue(executor, ginjects.ProvideWithPriorityTop1())
	hub.defaultExecutor = executor
}

func Clear() {
	hub.clear()
}

func execute(ctx context.Context, event interface{}, o *Option) error {
	eventType := reflect.TypeOf(event)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}
	es, defaultExecute := hub.findExecutes(eventType)
	if len(es) == 0 && o.mustHaveSubscriber {
		return fmt.Errorf("gevents: event type `%T`, not find executor", event)
	}
	if len(es) == 0 {
		if defaultExecute == nil {
			return fmt.Errorf("gevents: event type `%T`, not find executor", event)
		} else {
			return defaultExecute.Execute(ctx, event)
		}
	}
	var err error
	for _, e := range es {
		err = e.Execute(ctx, event)
		if err != nil {
			break
		}
	}
	return err
}

var hub = &_hub{}

type _hub struct {
	executes        map[reflect.Type][]Executor
	defaultExecutor Executor
}

func (h *_hub) register(executors ...Executor) {
	if h.executes == nil {
		h.executes = map[reflect.Type][]Executor{}
	}
	for _, executor := range executors {
		ginjects.ProvideByValue(executor, ginjects.ProvideWithPriorityTop1())
		for _, eventType := range executor.Types() {
			if eventType.Kind() == reflect.Ptr {
				eventType = eventType.Elem()
			}
			h.executes[eventType] = append(h.executes[eventType], executor)
		}
	}
}
func (h *_hub) clear() {
	h.executes = nil
	h.defaultExecutor = nil
}

func (h *_hub) SetDefaultExecutor(executor Executor) {
	ginjects.ProvideByValue(executor)
	h.defaultExecutor = executor
}

func (h *_hub) findExecutes(eventType reflect.Type) ([]Executor, Executor) {
	return h.executes[eventType], h.defaultExecutor
}
