package quantum

import (
	"crypto/tls"
	"errors"
	"time"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux"
	"github.com/doubledutch/mux/gob"
)

var (
	// ErrInvalidLager = nil Lager
	ErrInvalidLager = errors.New("Invalid Config Lager")
	// ErrInvalidPool = nil Pool
	ErrInvalidPool = errors.New("Invalid Config Pool")
)

// Config represents components required by all quantum components
type Config struct {
	Lager lager.Lager
	Pool  mux.Pool
}

// Verify validates a Config, checking for nil components
func (c *Config) Verify() error {
	if c.Lager == nil {
		return ErrInvalidLager
	}

	if c.Pool == nil {
		return ErrInvalidPool
	}

	return nil
}

// DefaultConfig is the default config
func DefaultConfig() *Config {
	return &Config{
		Lager: lager.NewLogLager(nil),
		Pool:  new(gob.Pool),
	}
}

// ConnConfig are configuration settings needed for Conn
type ConnConfig struct {
	Timeout   time.Duration
	TLSConfig *tls.Config
	*Config
}

// DefaultConnConfig is the default ConnConfig
func DefaultConnConfig() *ConnConfig {
	return &ConnConfig{
		Timeout: 100 * time.Millisecond,
		Config:  DefaultConfig(),
	}
}

// ToMux creates a mux.Config from ConnConfig
func (c *ConnConfig) ToMux() *mux.Config {
	if c == nil {
		return mux.DefaultConfig()
	}

	return &mux.Config{
		Timeout: c.Timeout,
		Lager:   c.Lager,
	}
}
