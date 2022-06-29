package events

import (
	"context"
	"fmt"
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
}

func (h *DefaultEventHandler) Execute(ctx context.Context, event interface{}) error {
	fmt.Printf("default:%v", event)
	return nil
}

func (h *DefaultEventHandler) Types() []reflect.Type {
	return nil
}

type UserModifyEventHandler struct {
}

func (h *UserModifyEventHandler) Execute(ctx context.Context, event interface{}) error {
	fmt.Println(event.(*UserModifyEvent))
	return nil
}

func (h *UserModifyEventHandler) Types() []reflect.Type {
	return []reflect.Type{reflect.TypeOf(&UserModifyEvent{})}
}

func TestMustHaveSubscriber(t *testing.T) {
	Register(&UserModifyEventHandler{})
	SetDefaultExecutor(&DefaultEventHandler{})
	publisher := &DefaultPublisher{}
	err := publisher.Publish(context.Background(), &UserModifyEvent{
		Id:    0,
		Name:  "zhaoche",
		State: "add",
	}, MustHaveSubscriber())
	if err != nil {
		t.Fatal(err)
	}
	err = publisher.Publish(context.Background(), &OrderModifyEvent{
		Id:    0,
		Name:  "zhaoche",
		State: "add",
	})
	if err != nil {
		t.Fatal(err)
	}
}
