package client

import (
	"fmt"
	"net"

	"github.com/doubledutch/quantum"
)

// Client sends requests to Server and reads the response
type Client struct {
	*quantum.ConnConfig
}

// New returns a new Client with the specified host and port
func New(config *quantum.ConnConfig) quantum.Client {
	return &Client{
		ConnConfig: config,
	}
}

// Dial connects to the ClientConfig.addr and returns ClientConn
func (c *Client) Dial(address string) (quantum.ClientConn, error) {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial err: %s", err)
	}

	conn, err := NewConn(netConn, c.ConnConfig)

	return conn, err
}
