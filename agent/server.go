package agent

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/inmemory"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux/gob"
)

var (
	// ErrUnexpectedError describes the error when a job panics
	ErrUnexpectedError = errors.New("Job exited with unexpected error")
)

// Config encapsulates configuration for an agent
type Config struct {
	*quantum.ConnConfig

	Port string

	Registry    quantum.Registry
	Registrator quantum.Registrator
}

// Agent routes job requests to jobs, and runs the jobs with the request
type Agent struct {
	*quantum.ConnConfig

	port  string
	done  chan struct{}
	sigCh chan os.Signal

	quantum.Registry
	registrator quantum.Registrator
}

// New creates a new Agent with the specified port
func New(config *Config) quantum.Agent {
	if config.Pool == nil {
		config.Pool = new(gob.Pool)
	}

	if config.Lager == nil {
		config.Lager = lager.NewLogLager(&lager.LogConfig{
			Levels: lager.LevelsFromString("IE"),
			Output: os.Stdout,
		})
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

	fmt.Println(config)

	return &Agent{
		ConnConfig: config.ConnConfig,

		Registry:    config.Registry,
		registrator: config.Registrator,

		port:  config.Port,
		done:  make(chan struct{}),
		sigCh: make(chan os.Signal, 1),
	}
}

// Accept accepts on a specified net.Listener
func (a *Agent) Accept(ln net.Listener) error {
	netConn, err := ln.Accept()
	if err != nil {
		return err
	}

	conn, err := NewConn(netConn, a.ConnConfig)
	if err != nil {
		a.Lager.Errorf("Error creating agent conn: %s", err)
		return err
	}

	go conn.Serve(a)
	return nil
}

// IsShutdown returns a chan that determines whether we're shutdown
func (a *Agent) IsShutdown() chan struct{} {
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
	defer close(a.sigCh)

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
	if err := a.registrator.Register(NewPort(a.port).Int(), a); err != nil {
		a.Lager.Errorf("Failed to announce services: %s\n", err)
		return err
	}

	// Blocks
	a.Lager.Debugf("ListenAndServe block")
	return quantum.ListenAndServe(a, a.port, a.Lager)
}

// Port holds a port in the form :XXXX
type Port struct {
	Value string
}

// NewPort returns a Port
func NewPort(s string) Port {
	return Port{Value: s}
}

// Int returns a int representation of Port
func (p Port) Int() int {
	number := p.Value[1:] // ignore : at [0]

	i, err := strconv.Atoi(number)
	if err != nil {
		return -1
	}
	return i
}
