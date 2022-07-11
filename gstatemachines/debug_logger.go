package gstatemachines

import "fmt"

var isDebugOn = false

func DebugLog(message string) {
	if isDebugOn {
		fmt.Println(message)
	}
}

func enableDebug() {
	isDebugOn = true
}
