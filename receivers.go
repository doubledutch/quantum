package quantum

import (
	"github.com/doubledutch/mux"
	"github.com/doubledutch/mux/gob"
)

// RequestReceiver receives Request
type RequestReceiver struct {
	dec *mux.Decoder
	ch  chan Request
}

// NewRequestReceiver creates a new request receiver
func NewRequestReceiver(ch chan Request) mux.Receiver {
	return mux.NewReceiver(ch, new(gob.Pool))
}
