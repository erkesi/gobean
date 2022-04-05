package application

import "testing"

func Test_callbackOrders_sort(t *testing.T) {
	tests := []struct {
		name string
		cos  callbackOrders
	}{
		{
			name: "normal",
			cos: callbackOrders{&callbackOrder{
				index:    1,
				callback: nil,
			}, &callbackOrder{
				index:    0,
				callback: nil,
			}, &callbackOrder{
				index:    2,
				priority: 1,
				callback: nil,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cos.sort()
			t.Log(tt.cos)
		})
	}
}
