package inject

import (
	"reflect"
	"testing"
)

func TestToposort(t *testing.T) {
	type args struct {
		edges    []Edge
		allNodes []EdgeNode
	}
	tests := []struct {
		name    string
		args    args
		want    []EdgeNode
		want1   []EdgeNode
		wantErr bool
	}{
		{name: "正常", args: args{
			edges: []Edge{{EdgeNode{
				index:    7,
				priority: 100,
			}, EdgeNode{
				index:    13,
				priority: 0,
			}}, {EdgeNode{
				index:    8,
				priority: 0,
			}, EdgeNode{
				index:    13,
				priority: 0,
			}}},
			allNodes: []EdgeNode{{
				index:    15,
				priority: 1000,
			}, {
				index:    16,
				priority: 10,
			}},
		}, want: []EdgeNode{{
			index:    15,
			priority: 1000,
		}, {
			index:    7,
			priority: 100,
		}, {
			index:    16,
			priority: 10,
		}, {
			index:    8,
			priority: 0,
		}, {
			index:    13,
			priority: 0,
		}}}, {name: "循环", args: args{
			edges: []Edge{{EdgeNode{
				index:    7,
				priority: 0,
			}, EdgeNode{
				index:    13,
				priority: 0,
			}}, {EdgeNode{
				index:    13,
				priority: 0,
			}, EdgeNode{
				index:    8,
				priority: 0,
			}}, {EdgeNode{
				index:    8,
				priority: 0,
			}, EdgeNode{
				index:    7,
				priority: 0,
			}}},
		}, want1: []EdgeNode{{
			index:    7,
			priority: 0,
		}, {
			index:    8,
			priority: 0,
		}, {
			index:    13,
			priority: 0,
		}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, _ := Toposort(tt.args.edges, tt.args.allNodes)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Toposort() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Toposort() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
