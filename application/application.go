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
	callback        appStateCallback
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

type appStateCallback func()

var index int

var state2CallbackOrders = map[state]callbackOrders{}

// AddInitCallbackWithPriority  priority 越大越先调 callback
func AddInitCallbackWithPriority(priority int, callback appStateCallback) {
	index++
	addCallbacks(stateInit, &callbackOrder{
		index:    index,
		priority: priority,
		callback: callback,
	})
}

func AddInitCallbacks(callbacks ...appStateCallback) {
	var callbackOrders []*callbackOrder
	for _, appStateCallback := range callbacks {
		index++
		callbackOrders = append(callbackOrders, &callbackOrder{
			index:    index,
			callback: appStateCallback,
		})
	}
	addCallbacks(stateInit, callbackOrders...)
}

// AddCloseCallbackWithPriority  priority 越大越先调 callback
func AddCloseCallbackWithPriority(priority int, callback appStateCallback) {
	index++
	addCallbacks(stateClose, &callbackOrder{
		index:    index,
		priority: priority,
		callback: callback,
	})
}

func AddCloseCallbacks(callbacks ...appStateCallback) {
	var callbackOrders []*callbackOrder
	for _, callback := range callbacks {
		index++
		callbackOrders = append(callbackOrders, &callbackOrder{
			index:    index,
			callback: callback,
		})
	}
	addCallbacks(stateClose, callbackOrders...)
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
