package quantum

import (
	"errors"

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

// Config represents configuration for a quantum component
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
