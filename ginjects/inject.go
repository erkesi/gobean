package ginjects

import (
	"context"
	"fmt"
	"math"
	"reflect"

	"github.com/erkesi/gobean/glogs"
)

type ObjectInit interface {
	Init()
}

type ObjectClose interface {
	Close()
}

var g = &graph{refType2InjectFieldIndies: map[reflect.Type]map[int]struct{}{}}

func Init() {
	g.Do(func() {
		if err := g.populate(); err != nil {
			panic(err)
		}
		for _, obj := range g.objects() {
			initFn := obj.initFn
			if initFn != nil {
				if glogs.Log != nil {
					glogs.Log.Debugf(context.TODO(), "ginjects: init inject object(%v)", obj)
				}
				initFn()
			}
		}
	})
}

func Close() {
	for i := len(g.objects()) - 1; i >= 0; i-- {
		closeFn := g.objects()[i].closeFn
		if closeFn != nil {
			if glogs.Log != nil {
				glogs.Log.Debugf(context.TODO(), "ginjects: close inject object(%v)", g.objects()[i])
			}
			closeFn()
		}
	}
}

type ProvideOption func(opt *provideOptions)

// WithProvidePriority
// priority 越大越先初始化（在按照依赖顺序的前提下）
func WithProvidePriority(priority int) ProvideOption {
	return func(opt *provideOptions) {
		opt.priority = priority
	}
}

// WithProvidePriorityTop1 最先初始化
func WithProvidePriorityTop1() ProvideOption {
	return func(opt *provideOptions) {
		opt.priority = math.MaxInt64
	}
}

type provideOptions struct {
	priority int
}

func apply(opts ...ProvideOption) *provideOptions {
	opt := &provideOptions{}
	for _, fn := range opts {
		fn(opt)
	}
	return opt
}

// ProvideByName 通过命名注入实例
// @param name string "命名"
// @param value interface 实例
func ProvideByName(name string, value interface{}, opts ...ProvideOption) {
	opt := apply(opts...)
	if err := g.provide(&object{
		priority: opt.priority,
		value:    value,
		name:     name,
	}); err != nil {
		panic(err)
	}
}

// ObtainByName 通过命名获取实例
// @param name string "命名"
// @return 指定名字的实例
func ObtainByName(name string) interface{} {
	if obj, ok := g.named[name]; ok {
		return obj.value
	}
	panic(fmt.Sprintf("ginjects: not found name `%s` instance", name))
}

// ProvideByValue
// @description 通过实例类型注入实例
// @param value 实例
func ProvideByValue(value interface{}, opts ...ProvideOption) {
	opt := apply(opts...)
	if err := g.provide(&object{
		priority: opt.priority,
		value:    value,
	}); err != nil {
		panic(err)
	}
}

// ObtainByType 通过值的类型（可以是interface或struct）获取实例
// @param value interface "值（类型可以是interface或struct）"
// @return 匹配值的类型的实例
func ObtainByType(value interface{}) interface{} {
	typ := reflect.TypeOf(value)
	if typ == nil {
		panic("ginjects: the value must be a reference pointer")
	}
	var realVal interface{}
	if typ.Kind() == reflect.Ptr {
		elemTyp := typ.Elem()
		switch elemTyp.Kind() {
		case reflect.Interface:
			var assignableTypes []string
			for tp, val := range g.unnamedType {
				if tp.AssignableTo(elemTyp) {
					assignableTypes = append(assignableTypes, tp.String())
					realVal = val
				}
			}
			if realVal == nil {
				panic(fmt.Sprintf("ginjects: not found type `%s` instance or implement type `%s`", elemTyp.String(), elemTyp.String()))
			}
			if len(assignableTypes) > 1 {
				panic(fmt.Sprintf("ginjects: exist type `%v` instance or implement type `%s`", assignableTypes, elemTyp.String()))
			}
		case reflect.Struct:
			if val, ok := g.unnamedType[typ]; ok {
				realVal = val
			} else {
				panic(fmt.Sprintf("ginjects: not found type `%s` instance", typ.String()))
			}
		default:
			panic("ginjects: the value must be a reference pointer to struct or interface")
		}
	}
	if realVal != nil {
		return realVal
	}
	panic("ginjects: the value must be a reference pointer")
}

func PrintObjects() []string {
	var r []string
	for _, o := range g.objects() {
		r = append(r, o.String())
	}
	return r
}
