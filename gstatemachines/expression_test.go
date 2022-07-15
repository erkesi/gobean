package gstatemachines

import (
	"errors"
	"fmt"
	"testing"
)

func Test_testExpression(t *testing.T) {
	vars := make(map[string]interface{})
	vars["toTask1"] = true
	if got, _ := testExpression("toTask1 == true", vars); got != true {
		t.Errorf("testExpression() = %v, want %v", got, true)
	} else {
		fmt.Println("toTask1 == true")
	}
	vars["toTask2"] = "true"
	if got, _ := testExpression(`toTask2 == "true"`, vars); got != true {
		t.Errorf("testExpression() = %v, want %v", got, true)
	} else {
		fmt.Println(`toTask2 == true`)
	}
	if got, _ := testExpression(`toTask2 == true`, vars); got == true {
		t.Errorf("testExpression() = %v, want %v", got, true)
	} else {
		fmt.Println(`toTask2 == "true"`)
	}
}

func Test_warp_error(t *testing.T) {
	err1 := errors.New("err1")
	warpErr := fmt.Errorf("err2, err:%w", err1)
	fmt.Println(warpErr)
	fmt.Println(errors.Unwrap(warpErr) == err1)
	fmt.Println(errors.Is(warpErr, err1))
}
