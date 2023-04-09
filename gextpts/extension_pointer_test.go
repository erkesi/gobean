package gextpts

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type UserExtensionPointer0 struct {
}

func (e *UserExtensionPointer0) Match(ctx context.Context, values ...interface{}) bool {
	fmt.Println(values[0])
	return true
}

type UserExtensionPointer1 struct {
}

func (e *UserExtensionPointer1) Match(ctx context.Context, values ...interface{}) bool {
	return values[0].(*User).Id == 1
}

func (e *UserExtensionPointer1) Validate(ctx context.Context, user *User) (bool, error) {
	fmt.Println(ctx.Value("tmp"))
	return user.Id == 1, errors.New("validate error")
}

type UserExtensionPointer2 struct {
}

func (e *UserExtensionPointer2) Match(ctx context.Context, values ...interface{}) bool {
	return values[0].(*User).Id == 2
}

func (e *UserExtensionPointer2) Validate(ctx context.Context, user *User) (bool, error) {
	return user.Id == 2, nil
}

type User struct {
	Id int64
}

type DataValidateExtPt interface {
	ExtensionPointer
	Validate(ctx context.Context, user *User) (bool, error)
}

func TestInterfaceFunction2(t *testing.T) {
	Register(&UserExtensionPointer1{}, WithExtPtPriority(99))
	Register(&UserExtensionPointer2{}, WithExtPtPriority(98))
	Register(&UserExtensionPointer0{}, WithExtPtPriority(100))
	ctx := context.WithValue(context.Background(), "tmp", "tmp_val")
	_, b, err := ExecuteWithErr(ctx, DataValidateExtPt.Validate, &User{Id: 1})
	fmt.Println(b.(bool))
	fmt.Println(err)
	if err.Error() != "validate error" {
		t.Fatalf("actual:%s, expected:%s", err.Error(), "validate error")
	}
}

func TestInterfaceFunction(t *testing.T) {
	var f = DataValidateExtPt.Validate
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func {
		panic("not func")
	}
	fmt.Println(fn.Kind())
	fmt.Println(fn.Type().NumIn())
	fmt.Println(fn.Type().In(0))
	fmt.Println(fn.Type().In(1))
	fmt.Println(fn.Type().In(2))
	input := []reflect.Value{reflect.ValueOf(&UserExtensionPointer1{}),
		reflect.ValueOf(context.TODO()), reflect.ValueOf(&User{Id: 1})}
	matchF := ExtensionPointer.Match
	matchFn := reflect.ValueOf(matchF)
	rets := matchFn.Call(input)
	if rets[0].Bool() {
		rets := fn.Call(input)
		fmt.Println(rets[0].Interface())
	}
}
