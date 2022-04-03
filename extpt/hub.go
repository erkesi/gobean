package extpt

import (
	"fmt"
	"github.com/erkesi/gobean/inject"
	"reflect"
	"sort"
	"sync"
)

var Hub = &hub{typeSet: make(map[reflect.Type]bool)}

type ExtPt struct {
	t               reflect.Type
	val             ExtensionPointer
	index, priority int64
}

type ExtPtFunc func(opt *extPtOpt)

// ExtPtWithPriority
// priority 越大越先初始化（在按照依赖顺序的前提下）
func ExtPtWithPriority(priority int64) ExtPtFunc {
	return func(opt *extPtOpt) {
		opt.priority = priority
	}
}

type extPtOpt struct {
	priority int64
}

func extPtOptsExec(opts ...ExtPtFunc) *extPtOpt {
	opt := &extPtOpt{}
	for _, fn := range opts {
		fn(opt)
	}
	return opt
}

type hub struct {
	extPts  []*ExtPt
	typeSet map[reflect.Type]bool
	m       sync.Map
	index   int64
}

func (h *hub) Register(extPt ExtensionPointer, opts ...ExtPtFunc) {
	t := reflect.TypeOf(extPt)
	if h.typeSet[reflect.TypeOf(extPt)] {
		panic(fmt.Sprintf("ExtensionPointer type(%s) exist", t.String()))
	}
	h.index = h.index + 1
	opt := extPtOptsExec(opts...)
	inject.ProvideByValue(extPt, inject.ProvideWithPriority(opt.priority))
	h.typeSet[reflect.TypeOf(extPt)] = true
	h.extPts = append(h.extPts, &ExtPt{
		t:        t,
		val:      extPt,
		index:    h.index,
		priority: opt.priority,
	})
	sort.Slice(h.extPts, func(i, j int) bool {
		if h.extPts[i].priority == h.extPts[j].priority {
			return h.extPts[i].index < h.extPts[j].index
		}
		return h.extPts[i].priority > h.extPts[j].priority
	})
}

func (h *hub) find(ifaceType reflect.Type) []ExtensionPointer {
	if value, ok := h.m.Load(ifaceType); ok {
		return value.([]ExtensionPointer)
	}
	extPts := make([]ExtensionPointer, 0)
	for _, extPt := range h.extPts {
		if extPt.t.AssignableTo(ifaceType) {
			extPts = append(extPts, extPt.val)
		}
	}
	h.m.Store(ifaceType, extPts)
	return extPts
}
