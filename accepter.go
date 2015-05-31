package quantum

import (
	"errors"
	"io"
	"log"
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
	netaddr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		a.Close()
		return ErrInvalidAddr
	}
	ln, err := net.ListenTCP("tcp", netaddr)
	if err != nil {
		a.Close()
		return ErrListen
	}
	lgr.Infof("Listening on " + port)
RECV_LOOP:
	for {
		// Set a deadline so we can check for shutdown
		ln.SetDeadline(time.Now().Add(500 * time.Millisecond))
		select {
		case <-a.IsShutdown():
			a.Close()
			break RECV_LOOP
		default:
		}

		if err := a.Accept(ln); err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}

			log.Println(err)
		}
	}

	return nil
}
