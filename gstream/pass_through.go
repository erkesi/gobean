package gstream

// passThrough retransmits incoming elements as is.
//
// in  -- 1 -- 2 ---- 3 -- 4 ------ 5 --
//
// out -- 1 -- 2 ---- 3 -- 4 ------ 5 --
type passThrough struct {
	StateConf
	in  chan interface{}
	out chan interface{}
}

// Verify passThrough satisfies the Transfer interface.
var _ Transfer = (*passThrough)(nil)

// newPassThrough returns a new passThrough instance.
func newPassThrough() *passThrough {
	passThrough := &passThrough{
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}
	go passThrough.doStream()

	return passThrough
}

// Via streams data through the given flow
func (pt *passThrough) Via(flow Transfer) Transfer {
	flow.setState(pt.State())
	go pt.transmit(flow)
	return flow
}

// To streams data to the given sink
func (pt *passThrough) To(sink Sink) {
	sink.SetSinkState(pt.State())
	go pt.transmit(sink)
}

// Out returns an output channel for sending data
func (pt *passThrough) Out() <-chan interface{} {
	return pt.out
}

// In returns an input channel for receiving data
func (pt *passThrough) In() chan<- interface{} {
	return pt.in
}

func (pt *passThrough) transmit(inlet Inlet) {
	for elem := range pt.Out() {
		inlet.In() <- elem
	}
	close(inlet.In())
}

func (pt *passThrough) doStream() {
	for elem := range pt.in {
		pt.out <- elem
	}
	close(pt.out)
}