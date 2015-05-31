package agent

import (
	"errors"
	"log"
	"net"
	"os"
	"runtime/debug"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux"
	"github.com/doubledutch/quantum"
)

// Conn wraps a Connection
type Conn struct {
	mux.Server

	// Senders
	OutCh chan string

	// Receivers
	SigCh     chan os.Signal
	RequestCh chan quantum.Request

	lgr lager.Lager
}

// NewConn returns a new Connection connected to the specified io.ReadWriter
func NewConn(conn net.Conn, config *quantum.Config) (*Conn, error) {
	muxConfig := mux.DefaultConfig()
	if config == nil {
		config = quantum.DefaultConfig()
	}

	muxConfig.Lager = config.Lager
	srv, err := config.Pool.NewServer(conn, muxConfig)
	if err != nil {
		return nil, err
	}

	ac := &Conn{
		Server:    srv,
		OutCh:     make(chan string, 1),
		SigCh:     make(chan os.Signal, 1),
		RequestCh: make(chan quantum.Request, 1),

		lgr: config.Lager,
	}

	// Send up receiver for signals
	sigR := ac.Pool().NewReceiver(ac.SigCh)
	srv.Receive(mux.SignalType, sigR)

	requestR := quantum.NewRequestReceiver(ac.RequestCh)
	srv.Receive(quantum.RequestType, requestR)

	go srv.Recv()

	return ac, nil
}

// Signals returns the signals channel of the connection
func (conn *Conn) Signals() chan os.Signal {
	return conn.SigCh
}

// Logs returns the logs channel of the connection
func (conn *Conn) Logs() chan string {
	return conn.OutCh
}

// Serve processes a connection and serves the response
func (conn *Conn) Serve(reg quantum.Registry) {
	conn.Done(conn.serve(reg))
}

func (conn *Conn) serve(reg quantum.Registry) (err error) {
	var request quantum.Request
	select {
	case request = <-conn.RequestCh:
	case <-conn.IsShutdown():
		return errors.New("connection shutdown")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("job err: %s\n%s", r, debug.Stack())
			err = ErrUnexpectedError
		}
	}()

	conn.lgr.Debugf("Received request: %s, %s\n", request.Type, request.Data)

	job, err := reg.Get(request)
	if err != nil {
		return
	}

	err = job.Run(conn)
	conn.lgr.Infof("job completed: %s\n", err)
	return
}
