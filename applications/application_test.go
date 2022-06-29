package applications

import (
	"reflect"
	"testing"
)

func Test_Callback(t *testing.T) {

	var initCase []string
	var closeCase []string

	AddInitCallback(func() {
		initCase = append(initCase, "init callback 1")
	})
	AddInitCallback(func() {
		initCase = append(initCase, "init callback 2, Priority:100")
	}, CallbackWithPriority(100))
	AddInitCallback(func() {
		initCase = append(initCase, "init callback 3, Priority:99")
	}, CallbackWithPriority(99))

	AddCloseCallback(func() {
		closeCase = append(closeCase, "close callback 1")
	})
	AddCloseCallback(func() {
		closeCase = append(closeCase, "close callback 2, Priority:100")
	}, CallbackWithPriority(100))

	AddCloseCallback(func() {
		closeCase = append(closeCase, "close callback 3, Priority:99")
	}, CallbackWithPriority(99))

	// init
	Init()
	if !reflect.DeepEqual([]string{"init callback 2, Priority:100", "init callback 3, Priority:99", "init callback 1"}, initCase) {
		t.Fatal("init case bug")
	}

	// close
	Close()
	if !reflect.DeepEqual([]string{"close callback 2, Priority:100", "close callback 3, Priority:99", "close callback 1"}, closeCase) {
		t.Fatal("close case bug")
	}

}
