package gextpts

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type UserExtensionPointer0 struct {
}

func (e *UserExtensionPointer0) Match(values ...interface{}) bool {
	fmt.Println(values[0])
	return true
}

type UserExtensionPointer1 struct {
}

func (e *UserExtensionPointer1) Match(values ...interface{}) bool {
	fmt.Println(values[0])
	return true
}

func (e *UserExtensionPointer1) Validate(user *User) (bool, error) {
	return user.Id == 1, errors.New("validate error")
}

type UserExtensionPointer2 struct {
}

func (e *UserExtensionPointer2) Match(values ...interface{}) bool {
	fmt.Println(values[0])
	return true
}

func (e *UserExtensionPointer2) Validate(user *User) (bool, error) {
	return user.Id == 2, errors.New("validate error")
}

type User struct {
	Id int64
}

type DataValidateExtPt interface {
	ExtensionPointer
	Validate(user *User) (bool, error)
}

func TestInterfaceFunction2(t *testing.T) {
	Hub.Register(&UserExtensionPointer1{}, ExtPtWithPriority(99))
	Hub.Register(&UserExtensionPointer2{}, ExtPtWithPriority(98))
	Hub.Register(&UserExtensionPointer0{}, ExtPtWithPriority(100))
	_, b, err := ExecuteWithErr(DataValidateExtPt.Validate, &User{Id: 1})
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
	input := []reflect.Value{reflect.ValueOf(&UserExtensionPointer1{}), reflect.ValueOf(&User{Id: 1})}
	matchF := ExtensionPointer.Match
	matchFn := reflect.ValueOf(matchF)
	rets := matchFn.Call(input)
	if rets[0].Bool() {
		rets := fn.Call(input)
		fmt.Println(rets[0].Interface())
	}
}
