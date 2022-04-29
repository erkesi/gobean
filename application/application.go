package application

import (
	"fmt"
	"sort"
)

type state int8

const (
	stateInit  state = 0
	stateClose state = 100
)

type callbackOrder struct {
	index, priority int
	callback        Callback
}

func (c callbackOrder) String() string {
	return fmt.Sprintf("index:%d, priority:%d", c.index, c.priority)
}

type callbackOrders []*callbackOrder

func (cos callbackOrders) sort() {
	sort.Slice(cos, func(i, j int) bool {
		if cos[i].priority == cos[j].priority {
			return cos[i].index < cos[j].index
		}
		return cos[i].priority > cos[j].priority
	})
}

type Callback func()

var index int

var state2CallbackOrders = map[state]callbackOrders{}

// AddInitCallback init callback
func AddInitCallback(callback Callback, opts ...CallbackOptFunc) {
	opt := optsExec(opts...)
	index++
	addCallbacks(stateInit, &callbackOrder{
		index:    index,
		priority: opt.priority,
		callback: callback,
	})
}

// AddCloseCallback close callback
func AddCloseCallback(callback Callback, opts ...CallbackOptFunc) {
	opt := optsExec(opts...)
	index++
	addCallbacks(stateClose, &callbackOrder{
		index:    index,
		priority: opt.priority,
		callback: callback,
	})
}

func Init() {
	callback(stateInit)
}

func Close() {
	callback(stateClose)
}

func addCallbacks(s state, callbackOrders ...*callbackOrder) {
	if cos, ok := state2CallbackOrders[s]; ok {
		cos = append(cos, callbackOrders...)
		state2CallbackOrders[s] = cos
	} else {
		state2CallbackOrders[s] = callbackOrders
	}
}

func callback(s state) {
	if cos, ok := state2CallbackOrders[s]; ok {
		cos.sort()
		for _, co := range cos {
			co.callback()
		}
	}
}

type CallbackOptFunc func(opt *callbackOpt)

// CallbackWithPriority
// priority 越大越先初始化
func CallbackWithPriority(priority int) CallbackOptFunc {
	return func(opt *callbackOpt) {
		opt.priority = priority
	}
}

type callbackOpt struct {
	priority int
}

func optsExec(opts ...CallbackOptFunc) *callbackOpt {
	opt := &callbackOpt{}
	for _, fn := range opts {
		fn(opt)
	}
	return opt
}
