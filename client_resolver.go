package quantum

import "errors"

// TODO: Resolve with Consul or DNS

var (
	// ErrNoConfigs occurs when there are no configs to work with
	ErrNoConfigs = errors.New("no configs provided")
	// ErrNoAgents occurs when no agents responded to pings
	ErrNoAgents = errors.New("no agents responded")
)

// ClientResolver defines the interface for a resolver client.
type ClientResolver interface {
	Resolve(request ResolveRequest) (ClientConn, error)
}

// ResolveRequest describes the parameters for client resolution
type ResolveRequest struct {
	Agent string
	Type  string
}

// ClientResolverConfig defines the parameters for ClientResolver
type ClientResolverConfig struct {
	Config *Config
	Server string
}
