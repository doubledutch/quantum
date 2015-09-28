package client

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/doubledutch/quantum"
)

// Client sends requests to Server and reads the response
type Client struct {
	*quantum.ConnConfig
}

// New returns a new Client with the specified host and port
func New(config *quantum.ConnConfig) quantum.Client {
	if config == nil {
		config = quantum.DefaultConnConfig()
	}

	return &Client{
		ConnConfig: config,
	}
}

// Dial connects to the address and returns quantum.ClientConn
func (c *Client) Dial(address string) (quantum.ClientConn, error) {
	return c.dialWithDialer(new(net.Dialer), address)
}

// DialTimeout connects to the address and returns quantum.ClientConn, timing out
// after time
func (c *Client) DialTimeout(address string, time time.Duration) (quantum.ClientConn, error) {
	return c.dialWithDialer(&net.Dialer{Timeout: time}, address)
}

func (c *Client) dialWithDialer(d *net.Dialer, address string) (quantum.ClientConn, error) {
	var netConn net.Conn
	var err error

	if c.ConnConfig.TLSConfig != nil {
		netConn, err = tls.DialWithDialer(d, "tcp", address, c.ConnConfig.TLSConfig)
	} else {
		netConn, err = d.Dial("tcp", address)
	}

	if err != nil {
		return nil, err
	}

	return NewConn(netConn, c.ConnConfig)
}
