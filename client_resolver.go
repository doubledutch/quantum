package quantum

import (
	"errors"
	"fmt"
	"strings"
)

// TODO: Resolve with Consul or DNS

var (
	// ErrNoConfigs occurs when there are no configs to work with
	ErrNoConfigs = errors.New("no configs provided")
)

// IsNoAgentsErr returns whether this error is a no agents responded error
func IsNoAgentsErr(err error) bool {
	return strings.Contains(err.Error(), "no agents responded")
}

// NoAgentsErr creates an error for no agents that responded with type
func NoAgentsErr(t string) error {
	return fmt.Errorf("no agents with responded with type %s", t)
}

// NoAgentsWithNameErr creates an error for no agents that responded with type and name
func NoAgentsWithNameErr(t string, name string) error {
	return fmt.Errorf("no agents with responded with type %s and name %s", t, name)
}

// NoAgentsFromRequest creates an error for no agents that responded from a request
func NoAgentsFromRequest(request ResolveRequest) (err error) {
	if request.Agent == "" {
		err = NoAgentsErr(request.Type)
	} else {
		err = NoAgentsWithNameErr(request.Type, request.Agent)
	}

	return
}

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
