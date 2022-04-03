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
				Order:    1,
				Callback: nil,
			}, &callbackOrder{
				Order:    0,
				Callback: nil,
			}, &callbackOrder{
				Order:    2,
				Callback: nil,
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
