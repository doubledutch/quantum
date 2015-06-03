package inmemory

import (
	"errors"

	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/client"
)

// ClientResolverConfig defines
type ClientResolverConfig struct {
	Config      *quantum.Config
	Registrator *Registrator
}

// ClientResolver provides resolution for jobs that are defined on an agent.
type ClientResolver struct {
	client      quantum.Client
	registrator *Registrator
}

// NewClientResolver creates a new ClientResolver using a Registrator.
func NewClientResolver(config ClientResolverConfig) (quantum.ClientResolver, error) {
	client := client.New(config.Config)

	if config.Registrator == nil {
		return nil, errors.New("Registrator required")
	}

	return &ClientResolver{
		client:      client,
		registrator: config.Registrator,
	}, nil
}

// Resolve resolves a ResolveRequest using Registrator
func (r *ClientResolver) Resolve(request quantum.ResolveRequest) (quantum.ClientConn, error) {
	var addr string
	var ok bool
	if addr, ok = r.registrator.Jobs[request.Type]; !ok {
		return nil, quantum.ErrNoAgents
	}

	return r.client.Dial(addr)
}
