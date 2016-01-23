package agent

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/inmemory"
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

	port       string
	done       chan struct{}
	sigCh      chan os.Signal
	isShutdown bool

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

func (a *Agent) shutdown() {
	if a.isShutdown {
		return
	}
	a.isShutdown = true
	close(a.done)
}

// Close shutdowns an agent
func (a *Agent) Close() error {
	a.shutdown()
	close(a.sigCh)
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

	go func() {
		sig := <-a.sigCh
		// If sig is nil, channel was closed
		if sig != nil {
			a.shutdown()
		}
	}()

	a.Lager.Debugf("Registering")
	if err := a.registrator.Register(NewPort(a.port).Int(), a); err != nil {
		a.Lager.Errorf("Failed to announce services: %s\n", err)
		return err
	}

	// Blocks
	a.Lager.Debugf("ListenAndServe block")
	// When agent shutdowns, ListenAndServe will Close the agents and return
	return quantum.ListenAndServe(a, a.port, a.Lager)
}
