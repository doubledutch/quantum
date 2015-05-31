package defaults

import (
	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/consul"
	"github.com/doubledutch/quantum/inmemory"
)

// NewClientResolver creates a default client resolver
func NewClientResolver(config quantum.ClientResolverConfig) quantum.ClientResolver {
	return consul.NewClientResolver(config)
}

// NewRegistry returns the default registry
func NewRegistry() quantum.Registry {
	return inmemory.NewRegistry()
}
