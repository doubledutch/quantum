package consul

import (
	"strconv"
	"strings"

	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/client"
	"github.com/doubledutch/lager"
	"github.com/miekg/dns"
)

type resolveResult struct {
	address string
}

// NewClientResolver creates a consul client resolver
func NewClientResolver(config quantum.ClientResolverConfig) quantum.ClientResolver {
	if config.Config == nil {
		config.Config = quantum.DefaultConfig()
	}
	return &ClientResolver{
		lgr:    config.Config.Lager,
		server: config.Server,
	}
}

// ClientResolver is a client resolver that leverages Consul's service discovery.
type ClientResolver struct {
	config *quantum.Config
	lgr    lager.Lager
	server string
}

// Resolve resolves a ClientConn using a ResolveRequest
func (cr *ClientResolver) Resolve(request quantum.ResolveRequest) (quantum.ClientConn, error) {
	// Get ResolveResults
	results, err := cr.resolveResults(request)
	if err != nil {
		return nil, err
	}

	// Ping each one, return the first one to respond
	return cr.resolveClient(results)
}

// ResolveConfigs resolves client configs given the specified arguments
func (cr *ClientResolver) resolveResults(rr quantum.ResolveRequest) (results []resolveResult, err error) {
	m := new(dns.Msg)
	srv := rr.Type + ".service.consul."
	m.SetQuestion(srv, dns.TypeSRV)

	c := &dns.Client{Net: "tcp"}
	in, _, err := c.Exchange(m, cr.server)
	if err != nil {
		cr.lgr.Errorf("DNS Exchange failed: %s\n", err)
		return nil, err
	}

	return newResolveResults(in, rr), nil
}

func newResolveResults(in *dns.Msg, rr quantum.ResolveRequest) (results []resolveResult) {
	for i, a := range in.Answer {
		srv := a.(*dns.SRV)
		if rr.Agent != "" {
			// We were given an agent name, match it to the hostname of the SRV record
			targetSplit := strings.Split(srv.Target, ".")
			if len(targetSplit) < 5 {
				// We expect name.node.dc1.consul. (5)
				continue
			}
			// Drop .node.dc1.consul.
			hostnameSplit := targetSplit[:len(targetSplit)-4]
			hostname := strings.Join(hostnameSplit, ".")
			if hostname != rr.Agent {
				continue
			}
		}
		a := in.Extra[i].(*dns.A)
		sPort := ":" + strconv.Itoa(int(srv.Port))
		results = append(results, resolveResult{
			address: a.A.String() + sPort,
		})
	}
	return
}

func (cr *ClientResolver) resolveClient(results []resolveResult) (conn quantum.ClientConn, err error) {
	err = quantum.ErrNoAgents
	// TODO: Do this concurrently
	for _, result := range results {
		client := client.New(&client.Config{
			Config: cr.config,
		})
		conn, err = client.Dial(result.address)
		if err != nil {
			break
		}
	}
	return
}
