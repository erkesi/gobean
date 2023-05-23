package gdataflow

type dataSource struct {
	FlowState
	in <-chan interface{}
}

func NewDataSource(out Outlet) Source {
	state := &State{}
	out.SetState(state)
	return &dataSource{
		FlowState: FlowState{state: state},
		in:        out.Out(),
	}
}

func (ds *dataSource) Out() <-chan interface{} {
	return ds.in
}

func (ds *dataSource) Via(flow Flow) Flow {
	flow.SetState(ds.State())
	doStream(ds, flow)
	return flow
}
