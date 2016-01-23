package quantum

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/doubledutch/lager"
)

var (
	// ErrInvalidAddr = error resolving TCP addr
	ErrInvalidAddr = errors.New("Invalid Address")
	// ErrListen = error listening on TCP port
	ErrListen = errors.New("Unable to listen on specified port")
)

// Acceptor defines an interface for accepting requests from a net.Listener
type Acceptor interface {
	Accept(net.Listener) error
	IsShutdown() chan struct{}
	io.Closer
}

// ListenAndServe is a common function for listening on a port and accepting connections
func ListenAndServe(a Acceptor, port string, lgr lager.Lager) error {
	defer a.Close()

	netaddr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return ErrInvalidAddr
	}
	ln, err := net.ListenTCP("tcp", netaddr)
	if err != nil {
		return ErrListen
	}
	lgr.Infof("Listening on %s", port)
RECV_LOOP:
	for {
		// Set a deadline so we can check for shutdown
		ln.SetDeadline(time.Now().Add(500 * time.Millisecond))
		select {
		case <-a.IsShutdown():
			break RECV_LOOP
		default:
		}

		if err := a.Accept(ln); err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}

			lgr.Errorf("Accepter accept err: %s", err)
		}
	}

	return nil
}
