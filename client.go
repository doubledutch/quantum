package quantum

import (
	"os"
	"time"

	"github.com/doubledutch/mux"
)

// Client creates ClientConns
type Client interface {
	Dial(address string) (ClientConn, error)
	DialTimeout(address string, time time.Duration) (ClientConn, error)
}

// ClientConn is a connection to an AgentConn
type ClientConn interface {
	mux.Client
	Run(request Request) error
	Logs() <-chan string
	Signals() chan<- os.Signal
}
