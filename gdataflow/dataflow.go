package gdataflow

import (
	"context"
	"sync"
)

type dataSource struct {
	TransStateConf
	in <-chan interface{}
}

func NewDataSource(ctx context.Context, out Outlet) Source {
	state := newTransState(ctx)
	out.SetTransState(state)
	return &dataSource{
		TransStateConf: TransStateConf{transState: state},
		in:             out.Out(),
	}
}

func (ds *dataSource) Out() <-chan interface{} {
	return ds.in
}

func (ds *dataSource) Via(flow Flow) Flow {
	flow.SetTransState(ds.TransState())
	ds.doStream(ds, flow)
	return flow
}

// doStream streams data from the outlet to inlet.
func (ds *dataSource) doStream(outlet Outlet, inlet Inlet) {
	go func() {
		for element := range outlet.Out() {
			inlet.In() <- element
		}
		close(inlet.In())
	}()
}

// Split splits the stream into two flows according to the given boolean predicate.
func Split[T any](outlet Outlet, predicate func(T) bool) [2]Flow {
	condTrue := newPassThrough()
	condTrue.SetTransState(outlet.TransState())
	condFalse := newPassThrough()
	condFalse.SetTransState(outlet.TransState())
	go func() {
		for element := range outlet.Out() {
			if predicate(element.(T)) {
				condTrue.In() <- element
			} else {
				condFalse.In() <- element
			}
		}
		close(condTrue.In())
		close(condFalse.In())
	}()

	return [...]Flow{condTrue, condFalse}
}

// FanOut creates a number of identical flows from the single outlet.
// This can be useful when writing to multiple sinks is required.
func FanOut(outlet Outlet, magnitude int) []Flow {
	var out []Flow
	for i := 0; i < magnitude; i++ {
		pt := newPassThrough()
		pt.SetTransState(outlet.TransState())
		out = append(out, pt)
	}

	go func() {
		for element := range outlet.Out() {
			for _, socket := range out {
				socket.In() <- element
			}
		}
		for i := 0; i < magnitude; i++ {
			close(out[i].In())
		}
	}()

	return out
}

// Merge merges multiple flows into a single flow.
func Merge(outlets ...Flow) Flow {
	merged := newPassThrough()
	merged.SetTransState(outlets[0].TransState())
	var wg sync.WaitGroup
	wg.Add(len(outlets))

	for _, out := range outlets {
		go func(outlet Outlet) {
			for element := range outlet.Out() {
				merged.In() <- element
			}
			wg.Done()
		}(out)
	}

	// close the in channel on the last outlet close.
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(merged.In())
	}(&wg)

	return merged
}
