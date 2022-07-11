package gstatemachines

import (
	"fmt"
	"testing"
)

func Test_testExpression(t *testing.T) {

	vars := make(map[string]interface{})
	vars["toTask1"] = true
	if got := testExpression("toTask1 == true", vars); got != true {
		t.Errorf("testExpression() = %v, want %v", got, true)
	} else {
		fmt.Println("toTask1 == true")
	}

	vars["toTask2"] = "true"
	if got := testExpression(`toTask2 == "true"`, vars); got != true {
		t.Errorf("testExpression() = %v, want %v", got, true)
	} else {
		fmt.Println(`toTask2 == true`)
	}
	if got := testExpression(`toTask2 == true`, vars); got != true {
		t.Errorf("testExpression() = %v, want %v", got, true)
	} else {
		fmt.Println(`toTask2 == "true"`)
	}
}
