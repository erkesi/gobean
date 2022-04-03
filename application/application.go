package application

import (
	"sort"
)

type state int8

const (
	stateInit  state = 0
	stateClose state = 100
)

type callbackOrder struct {
	Order    int // order越小越先执行
	Callback appStateCallback
}

type callbackOrders []*callbackOrder

func (cos callbackOrders) sort() {
	sort.Slice(cos, func(i, j int) bool {
		return cos[i].Order-cos[j].Order < 0
	})
}

type appStateCallback func()

var state2CallbackOrders = map[state]callbackOrders{}

func AddInitCallbackOrder(order int, callback appStateCallback) {
	addCallbackOrders(stateInit, &callbackOrder{
		Order:    order,
		Callback: callback,
	})
}

func AddInitCallbacks(callbacks ...appStateCallback) {
	var callbackOrders []*callbackOrder
	for _, appStateCallback := range callbacks {
		callbackOrders = append(callbackOrders, &callbackOrder{
			Order:    0,
			Callback: appStateCallback,
		})
	}
	addCallbackOrders(stateInit, callbackOrders...)
}

func AddCloseCallbackOrder(order int, callback appStateCallback) {
	addCallbackOrders(stateClose, &callbackOrder{
		Order:    order,
		Callback: callback,
	})
}

func AddCloseCallbacks(callbacks ...appStateCallback) {
	callbackOrders := toCallbackOrders(callbacks)
	addCallbackOrders(stateClose, callbackOrders...)
}

func toCallbackOrders(callbacks []appStateCallback) []*callbackOrder {
	var callbackOrders []*callbackOrder
	for _, callback := range callbacks {
		callbackOrders = append(callbackOrders, &callbackOrder{
			Order:    0,
			Callback: callback,
		})
	}
	return callbackOrders
}

func Init() {
	callback(stateInit)
}

func Close() {
	callback(stateClose)
}

func addCallbackOrders(s state, callbackOrders ...*callbackOrder) {
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
			co.Callback()
		}
	}
}
