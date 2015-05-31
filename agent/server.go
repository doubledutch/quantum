package agent

import (
	"errors"
	"flag"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/consul"
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
	*quantum.Config
	Port   string
	Server string

	Levels string

	Registry    quantum.Registry
	Registrator quantum.Registrator
}

// RegisterFlags takes defaultFlags and registers quantum agent flags
func RegisterFlags(defaultFlags *Config) *Config {
	config := new(Config)
	config.Config = quantum.DefaultConfig()
	if defaultFlags == nil {
		defaultFlags = new(Config)
	}
	if defaultFlags.Config == nil {
		defaultFlags.Config = quantum.DefaultConfig()
	}

	if defaultFlags.Levels == "" {
		defaultFlags.Levels = "IE"
	}

	flag.StringVar(&config.Port, "p", defaultFlags.Port, "port to run on")
	// This is required to create the Registrator
	flag.StringVar(&config.Server, "s", defaultFlags.Server, "quantum service server to register with")
	flag.StringVar(&config.Levels, "log", defaultFlags.Levels, "log levels")

	return config
}

// Agent routes job requests to jobs, and runs the jobs with the request
type Agent struct {
	config *quantum.Config

	port  string
	done  chan struct{}
	sigCh chan os.Signal

	quantum.Registry
	registrator quantum.Registrator

	lgr lager.Lager
}

// New creates a new Agent with the specified port
func New(config *Config) quantum.Agent {
	lager := lager.NewLogLager(&lager.LogConfig{
		Levels: lager.LevelsFromString(config.Levels),
		Output: os.Stdout,
	})

	if config.Config == nil {
		config.Config = &quantum.Config{
			Pool:  new(gob.Pool),
			Lager: lager,
		}
	}

	if config.Registry == nil {
		config.Registry = inmemory.NewRegistry()
	}

	if config.Registrator == nil {
		config.Registrator = consul.NewRegistrator(config.Server, lager)
	}

	return &Agent{
		config: config.Config,
		port:   config.Port,
		done:   make(chan struct{}),
		sigCh:  make(chan os.Signal, 1),

		// These should be injected as defaults
		Registry:    config.Registry,
		registrator: config.Registrator,

		lgr: lager,
	}
}

// Accept accepts on a specified net.Listener
func (a *Agent) Accept(ln net.Listener) error {
	netConn, err := ln.Accept()
	if err != nil {
		return err
	}

	conn, err := NewConn(netConn, a.config)
	if err != nil {
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

	if err := a.registrator.Register(NewPort(a.port).Int(), a); err != nil {
		a.lgr.Warnf("Failed to announce services: %s\n", err)
	}

	// Blocks
	return quantum.ListenAndServe(a, a.port, a.lgr)
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
