package client

import (
	"net"
	"os"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux"
	"github.com/doubledutch/quantum"
)

// Conn implements quantum.ClientConn
type Conn struct {
	mux.Client

	lgr lager.Lager

	logCh chan string
	sigCh chan os.Signal
}

// NewConn returns a new Connection connected to the specified io.ReadWriter
func NewConn(conn net.Conn, config *quantum.ConnConfig) (quantum.ClientConn, error) {
	if config == nil {
		config = quantum.DefaultConnConfig()
	}

	client, err := config.Pool.NewClient(conn, config.ToMux())

	if err != nil {
		return nil, err
	}
	cc := &Conn{
		Client: client,
		lgr:    config.Lager,
		logCh:  make(chan string, 1),
		sigCh:  make(chan os.Signal, 1),
	}

	// Send up receiver for logs
	logR := cc.Pool().NewReceiver(cc.logCh)
	client.Receive(mux.LogType, logR)

	go client.Recv()

	return cc, nil
}

// Logs provides the logs that the client receives
func (c *Conn) Logs() <-chan string {
	return c.logCh
}

// Signals provides a way to send signals to the other end
func (c *Conn) Signals() chan<- os.Signal {
	return c.sigCh
}

// Run sends the Request to the server on the other send
// and waits for the response.
func (c *Conn) Run(request quantum.Request) error {
	// Connections are single use
	defer c.Close()

	c.lgr.Debugf("Sending request: %s", request)
	if err := c.Send(quantum.RequestType, request); err != nil {
		c.lgr.Errorf("Error sending request: %s", err)
		return err
	}

	go func() {
		// Listen for signal, if occurs, send it to server
		for {
			sig := <-c.sigCh
			if sig != nil {
				c.Send(mux.SignalType, sig)
			} else {
				break
			}
		}
	}()

	c.lgr.Debugf("Waiting")
	err := c.Wait()
	return err
}

// Close closes ClientConn
func (c *Conn) Close() error {
	// We need to close senders, receivers are closed by mux.Client
	c.Client.Shutdown()
	close(c.sigCh)
	return nil
}
