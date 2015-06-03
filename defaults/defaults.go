package defaults

import (
	"os"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/consul"
	"github.com/doubledutch/quantum/inmemory"
)

// NewClientResolver creates a default client resolver
func NewClientResolver(config quantum.ClientResolverConfig) quantum.ClientResolver {
	return consul.NewClientResolver(config)
}

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
