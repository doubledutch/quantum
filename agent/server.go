package agent

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/inmemory"
)

var (
	// ErrUnexpectedError describes the error when a job panics
	ErrUnexpectedError = errors.New("Job exited with unexpected error")
)

// Config encapsulates configuration for an agent
type Config struct {
	*quantum.ConnConfig

	Addr string

	Registry    quantum.Registry
	Registrator quantum.Registrator
}

// Agent routes job requests to jobs, and runs the jobs with the request
type Agent struct {
	*quantum.ConnConfig

	addr  string
	done  chan struct{}
	sigCh chan os.Signal

	quantum.Registry
	registrator quantum.Registrator
}

// New creates a new Agent with the specified port
func New(config *Config) quantum.Agent {
	if config == nil {
		config = new(Config)
	}

	if config.ConnConfig == nil {
		config.ConnConfig = quantum.DefaultConnConfig()
	}

	if config.Registry == nil {
		config.Registry = inmemory.NewRegistry(config.Lager)
	}

	if config.Registrator == nil {
		config.Registrator = inmemory.NewRegistrator()
	}

	if config.Timeout == 0 {
		config.Timeout = 100 * time.Millisecond
	}

	return &Agent{
		ConnConfig: config.ConnConfig,

		Registry:    config.Registry,
		registrator: config.Registrator,

		addr:  config.Addr,
		done:  make(chan struct{}),
		sigCh: make(chan os.Signal, 1),
	}
}

// Serve serves a connection
func (a *Agent) Serve(netConn net.Conn) error {
	conn, err := NewConn(netConn, a.ConnConfig)
	if err != nil {
		a.Lager.Errorf("Error creating agent conn: %s", err)
		return err
	}

	go conn.Serve(a)
	return nil
}

// IsClosed return a chan determining if agent is closed or not
func (a *Agent) IsClosed() chan struct{} {
	return a.done
}

// Close shutdowns an agent
func (a *Agent) Close() error {
	// Deregister services
	return a.registrator.Deregister()
}

// Start starts the agent by setting up the signal listener and listening on port.
func (a *Agent) Start() error {
	// Listen for signals, wire up done

	signal.Notify(a.sigCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL)

	// TODO: We should also notify any running jobs
	// Or, we can start a goroutine to update each connection when this happens?
	go func() {
		sig := <-a.sigCh
		// If sig is nil, channel was closed
		if sig != nil {
			close(a.done)
		}
	}()

	a.Lager.Debugf("Registering")
	if err := a.registrator.Register(NewPort(a.addr).Port(), a); err != nil {
		a.Lager.Errorf("Failed to announce services: %s\n", err)
		return err
	}

	// Blocks
	if a.ConnConfig.TLSConfig != nil {
		a.Lager.Debugf("ListenAndServeTLS blocking")
		return quantum.ListenAndServeTLS(a, a.addr, a.ConnConfig.TLSConfig, a.Lager)
	}
	a.Lager.Debugf("ListenAndServe blocking")
	return quantum.ListenAndServe(a, a.addr, a.Lager)
}

// Port holds a network address
type Port struct {
	Value string
}

// NewPort returns a Port
func NewPort(s string) Port {
	return Port{Value: s}
}

// Port returns a int representation of Port
func (p Port) Port() int {
	netAddr, err := net.ResolveTCPAddr("tcp", p.Value)
	if err != nil {
		return -1
	}
	return netAddr.Port
}
