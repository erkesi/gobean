package gextpts

import (
	"fmt"
	"github.com/erkesi/gobean/ginjects"
	"reflect"
	"sort"
	"sync"
)

var Hub = &hub{typeSet: make(map[reflect.Type]bool)}

type extPt struct {
	t               reflect.Type
	val             ExtensionPointer
	index, priority int
}

type ExtPtOptFunc func(opt *extPtOpt)

// ExtPtWithPriority
// priority 越大越先初始化（在按照依赖顺序的前提下）
func ExtPtWithPriority(priority int) ExtPtOptFunc {
	return func(opt *extPtOpt) {
		opt.priority = priority
	}
}

type extPtOpt struct {
	priority int
}

func extPtOptsExec(opts ...ExtPtOptFunc) *extPtOpt {
	opt := &extPtOpt{}
	for _, fn := range opts {
		fn(opt)
	}
	return opt
}

type hub struct {
	extPts  []*extPt
	typeSet map[reflect.Type]bool
	m       sync.Map
	index   int
}

func (h *hub) Register(et ExtensionPointer, opts ...ExtPtOptFunc) {
	t := reflect.TypeOf(et)
	if h.typeSet[reflect.TypeOf(et)] {
		panic(fmt.Sprintf("ExtensionPointer type(%s) exist", t.String()))
	}
	h.index = h.index + 1
	opt := extPtOptsExec(opts...)
	ginjects.ProvideByValue(et, ginjects.ProvideWithPriority(opt.priority))
	h.typeSet[reflect.TypeOf(et)] = true
	h.extPts = append(h.extPts, &extPt{
		t:        t,
		val:      et,
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
