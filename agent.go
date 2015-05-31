package quantum

import (
	"os"

	"github.com/doubledutch/mux"
)

// Agent routes job requests to jobs, and runs the jobs with the request
type Agent interface {
	Acceptor
	Registry
	Start() error
}

// AgentConn is a connection created on an agent
type AgentConn interface {
	mux.Server
	Logs() chan string
	Signals() chan os.Signal
}
