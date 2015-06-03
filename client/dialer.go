package client

import (
	"fmt"
	"net"

	"github.com/doubledutch/quantum"
)

// Client sends requests to Server and reads the response
type Client struct {
	config *quantum.Config
}

// New returns a new Client with the specified host and port
func New(config *quantum.Config) quantum.Client {
	if config == nil {
		config = quantum.DefaultConfig()
	}

	return &Client{
		config: config,
	}
}

// Dial connects to the ClientConfig.addr and returns ClientConn
func (c *Client) Dial(address string) (quantum.ClientConn, error) {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial err: %s", err)
	}

	conn, err := NewConn(netConn, c.config)

	return conn, err
}
