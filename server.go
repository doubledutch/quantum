package quantum

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	"github.com/doubledutch/lager"
)

var (
	// ErrInvalidAddr = error resolving TCP addr
	ErrInvalidAddr = errors.New("Invalid Address")
	// ErrListen = error listening on TCP port
	ErrListen = errors.New("Unable to listen on specified port")
)

// Server serves net connections
type Server interface {
	Serve(net.Conn) error
	IsClosed() chan struct{}
	io.Closer
}

// ListenAndServe listens on an address and serves an Acceptor
func ListenAndServe(srv Server, addr string, lgr lager.Lager) error {
	tc, err := Listen(addr)
	if err != nil {
		return err
	}

	return Serve(tc, srv, lgr)
}

// ListenAndServeTLS listens on an address using a TLS configuration and serves Server
func ListenAndServeTLS(srv Server, addr string, config *tls.Config, lgr lager.Lager) error {
	tc, err := ListenTLS(addr, config)
	if err != nil {
		return err
	}

	return Serve(tc, srv, lgr)
}

// Listen listens on an address
func Listen(addr string) (net.Listener, error) {
	netaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, ErrInvalidAddr
	}
	tc, err := net.ListenTCP("tcp", netaddr)
	if err != nil {
		return nil, err
	}

	// Set a deadline so we can check for shutdown
	tc.SetDeadline(time.Now().Add(500 * time.Millisecond))
	return tcpWrapperListener{
		TCPListener: tc,
		KeepAlive:   3 * time.Minute,
		Deadline:    1 * time.Second,
	}, nil
}

// ListenTLS listens on an address with a TLS config
func ListenTLS(addr string, config *tls.Config) (net.Listener, error) {
	ln, err := Listen(addr)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(ln, config), nil
}

// Serve serves an Acceptor using a listener. Serve blocks.
func Serve(ln net.Listener, srv Server, lgr lager.Lager) error {
	lgr.Infof("Listening on %s", ln.Addr())
LOOP:
	for {
		conn, err := ln.Accept()
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "closed") || strings.Contains(err.Error(), "reset by peer") {
				// This is the expected way for us to return
				lgr.Infof("Accept loop disconnected from %s", conn.RemoteAddr())
				continue
			}
			if err, ok := err.(*net.OpError); ok && err.Timeout() {
				select {
				case <-srv.IsClosed():
					lgr.Debugf("Accept loop closing")
					break LOOP
				default: // Keep listening
					continue
				}
			} else {
				lgr.Errorf("Unexpected net.OpError: %s", err)
				continue
			}
		}

		if err := srv.Serve(conn); err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}

			lgr.Errorf("Server serve err: %s", err)
		}
	}

	return nil
}

type tcpWrapperListener struct {
	*net.TCPListener
	KeepAlive time.Duration
	Deadline  time.Duration
}

func (ln tcpWrapperListener) Accept() (c net.Conn, err error) {
	ln.SetDeadline(time.Now().Add(ln.Deadline))
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(ln.KeepAlive)
	return tc, nil
}
