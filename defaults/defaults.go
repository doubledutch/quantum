package defaults

import (
	"os"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/inmemory"
)

// NewRegistry returns a default registry
func NewRegistry() quantum.Registry {
	return inmemory.NewRegistry(NewLager())
}

// NewLager returns a default Lager
func NewLager() lager.Lager {
	return lager.NewLogLager(&lager.LogConfig{
		Levels: lager.LevelsFromString("IE"),
		Output: os.Stdout,
	})
}
