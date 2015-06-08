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

// MultiClientResolver implements ClientResolver by trying to resolve a client
// be iterating through provided ClientResolvers.
type MultiClientResolver struct {
	Resolvers []ClientResolver
}

// Resolve resolves a ResolveRequest by iterating through r.Resolvers
func (r *MultiClientResolver) Resolve(request ResolveRequest) (conn ClientConn, err error) {
	for _, resolver := range r.Resolvers {
		conn, err = resolver.Resolve(request)
		if err == nil {
			break
		}
	}

	return conn, err
}
