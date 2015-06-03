package client

import (
	"net"
	"os"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux"
	"github.com/doubledutch/quantum"
)

// Conn wraps a Connection
type Conn struct {
	mux.Client

	lgr lager.Lager

	LogCh chan string
	SigCh chan os.Signal
}

// NewConn returns a new Connection connected to the specified io.ReadWriter
func NewConn(conn net.Conn, config *quantum.Config) (*Conn, error) {
	muxConfig := mux.DefaultConfig()
	if config == nil {
		config = quantum.DefaultConfig()
	}

	muxConfig.Lager = config.Lager

	client, err := config.Pool.NewClient(conn, muxConfig)

	if err != nil {
		return nil, err
	}
	cc := &Conn{
		Client: client,
		lgr:    config.Lager,
		LogCh:  make(chan string, 1),
		SigCh:  make(chan os.Signal, 1),
	}

	// Send up receiver for logs
	logR := cc.Pool().NewReceiver(cc.LogCh)
	client.Receive(mux.LogType, logR)

	go client.Recv()

	config.Lager.Debugf("Connection initialized")

	return cc, nil
}

// Logs provides the logs that the client receives
func (c *Conn) Logs() <-chan string {
	return c.LogCh
}

// Signals provides a way to send signals to the other end
func (c *Conn) Signals() chan<- os.Signal {
	return c.SigCh
}

// Run sends the Request to the server on the other send
// and waits for the response.
func (c *Conn) Run(request quantum.Request) error {
	// Use type of request data to create a requester
	c.lgr.Debugf("Sending request: %s", request)
	if err := c.Send(quantum.RequestType, request); err != nil {
		c.lgr.Errorf("Error sending request: %s", request)
		return err
	}

	go func() {
		// Listen for signal, if occurs, send it to server
		for {
			sig := <-c.SigCh
			if sig != nil {
				c.Send(mux.SignalType, sig)
			} else {
				break
			}
		}
	}()

	c.lgr.Debugf("Waiting")
	err := c.Wait()
	c.Close()
	return err
}

// Close closes ClientConn
func (c *Conn) Close() error {
	// We need to close senders, receivers are closed by mux.Client
	close(c.SigCh)
	return nil
}
