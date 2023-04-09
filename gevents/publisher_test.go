package gevents

import (
	"context"
	"reflect"
	"testing"
)

type UserModifyEvent struct {
	Id    int
	Name  string
	State string
}

type OrderModifyEvent struct {
	Id    int
	Name  string
	State string
}

type DefaultEventHandler struct {
	T testing.TB
}

func (h *DefaultEventHandler) Execute(ctx context.Context, event interface{}) error {
	if e, ok := event.(*OrderModifyEvent); !ok || e.Id != 3 {
		h.T.Fatal("e.id != 3")
	}
	return nil
}

func (h *DefaultEventHandler) Types() []reflect.Type {
	return nil
}

type UserModifyEventHandler struct {
	T testing.TB
}

func (h *UserModifyEventHandler) Execute(ctx context.Context, event interface{}) error {
	if e, ok := event.(*UserModifyEvent); !ok || e.Id != 2 {
		h.T.Fatalf("e.id:%d", e.Id)
	}
	return nil
}

func (h *UserModifyEventHandler) Types() []reflect.Type {
	return []reflect.Type{reflect.TypeOf(&UserModifyEvent{})}
}

type UserModifyEventHandler1 struct {
	T testing.TB
}

func (h *UserModifyEventHandler1) Execute(ctx context.Context, event interface{}) error {
	if e, ok := event.(*UserModifyEvent); !ok || e.Id != 1 {
		h.T.Fatal("e.id != 1")
	} else {
		e.Id = 2
	}
	return nil
}

func (h *UserModifyEventHandler1) Types() []reflect.Type {
	return []reflect.Type{reflect.TypeOf(&UserModifyEvent{})}
}

func TestEventPublish(t *testing.T) {
	Register(&UserModifyEventHandler{T: t}, WithRegisterPriority(100))
	Register(&UserModifyEventHandler1{T: t}, WithRegisterPriority(101))
	SetDefaultExecutor(&DefaultEventHandler{T: t})
	publisher := &DefaultPublisher{}
	err := publisher.Publish(context.Background(), &UserModifyEvent{
		Id:    1,
		Name:  "zhaoche",
		State: "add",
	}, WithMustHaveSubscriber())
	if err != nil {
		t.Fatal(err)
		return
	}
	err = publisher.Publish(context.Background(), &OrderModifyEvent{
		Id:    3,
		Name:  "zhaoche",
		State: "add",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
}
