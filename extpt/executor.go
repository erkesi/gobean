package extpt

import (
	"fmt"
	"reflect"
)

// Execute 执行扩展点，接口方法返回值只有一个参数的情况
// @param f interface "接口方法"
// @param args []interface "接口方法参数"
// @return ok bool "是否匹配到了扩展点实例"
// @return value interface "接口方法返回值"
func Execute(f interface{}, args ...interface{}) (bool, interface{}) {
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func {
		panic("args[0] kind not func")
	}
	if fn.Type().NumOut() != 1 {
		panic(fmt.Sprintf("func `%v`, the number of returned parameters is not equal to 1", fn.Type()))
	}
	var inputArgs []reflect.Value
	for _, arg := range args {
		inputArgs = append(inputArgs, reflect.ValueOf(arg))
	}
	impls := find(fn)
	for _, impl := range impls {
		var input []reflect.Value
		input = append(input, reflect.ValueOf(impl))
		input = append(input, inputArgs...)
		matchF := ExtensionPointer.Match
		matchFn := reflect.ValueOf(matchF)
		rets := matchFn.Call(input)
		if rets[0].Bool() {
			rets := fn.Call(input)
			return true, rets[0].Interface()
		}
	}
	return false, nil
}

var e *error
var errorType = reflect.TypeOf(e)

// ExecuteWithErr 执行扩展点，接口方法返回值只有两个参数，第二个参数是Error接口类型
// @param f interface "接口方法"
// @param args []interface "接口方法参数"
// @return ok bool "是否匹配到了扩展点实例"
// @return value interface "接口方法返回第一个值"
// @return err Error "接口方法返回第二个值（Error类型）"
func ExecuteWithErr(f interface{}, args ...interface{}) (bool, interface{}, error) {
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func {
		panic("args[0] kind not func")
	}
	if fn.Type().NumOut() != 2 {
		panic(fmt.Sprintf("func `%v`, the number of returned parameters is not equal to 2", fn.Type()))
	}
	errType := fn.Type().Out(1)
	if errType != errorType.Elem() {
		panic(fmt.Sprintf("func `%v`, the second parameter returned is not of type `error`", fn.Type()))
	}
	var inputArgs []reflect.Value
	for _, arg := range args {
		inputArgs = append(inputArgs, reflect.ValueOf(arg))
	}
	impls := find(fn)
	for _, impl := range impls {
		var input []reflect.Value
		input = append(input, reflect.ValueOf(impl))
		input = append(input, inputArgs...)
		matchF := ExtensionPointer.Match
		matchFn := reflect.ValueOf(matchF)
		rets := matchFn.Call(input)
		if rets[0].Bool() {
			rets := fn.Call(input)
			var err error
			if !rets[1].IsZero() {
				err = rets[1].Interface().(error)
			}
			return true, rets[0].Interface(), err
		}
	}
	return false, nil, nil
}

func find(fn reflect.Value) []ExtensionPointer {
	t := fn.Type().In(0)
	if t.Kind() != reflect.Interface {
		panic(fmt.Sprintf("param f(%s), first param not is interface", fn.Type().String()))
	}
	impls := Hub.find(t)
	if len(impls) == 0 {
		panic(fmt.Sprintf("not find ExtensionPointer implement %s", t.String()))
	}
	return impls
}
