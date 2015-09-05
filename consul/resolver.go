package consul

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/client"
	"github.com/hashicorp/consul/api"
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
	config *quantum.ConnConfig
	lgr    lager.Lager
	server string

	dnsc  *dns.Client
	httpc *api.Client
}

// Resolve resolves a ClientConn using a ResolveRequest
func (cr *ClientResolver) Resolve(request quantum.ResolveRequest) (quantum.ClientConn, error) {
	// Get ResolveResults
	results, err := cr.resolveResults(request)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, quantum.NoAgentsFromRequest(request)
	}

	// Ping each one, return the first one to respond
	return cr.resolveClient(results)
}

// ResolveConfigs resolves client configs given the specified arguments
func (cr *ClientResolver) resolveResults(rr quantum.ResolveRequest) (results []resolveResult, err error) {
	if rr.Agent == "" {
		return cr.resolveWithDNS(rr)
	}

	return cr.resolveWithAPI(rr)
}

func (cr *ClientResolver) resolveWithDNS(rr quantum.ResolveRequest) (results []resolveResult, err error) {
	m := new(dns.Msg)
	// For now, assume .service.consul. for domain
	srv := rr.Type + ".service.consul."
	m.SetQuestion(srv, dns.TypeSRV)

	if cr.dnsc == nil {
		cr.dnsc = &dns.Client{Net: "tcp"}
	}
	// For now, assume :8600 for Consul DNS
	in, _, err := cr.dnsc.Exchange(m, cr.server+":8600")
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

func (cr *ClientResolver) resolveWithAPI(rr quantum.ResolveRequest) (results []resolveResult, err error) {
	if cr.httpc == nil {
		// For now, assume port :8500 for HTTP API
		cr.httpc, err = api.NewClient(&api.Config{
			Address: cr.server + ":8500",
		})
		if err != nil {
			return nil, err
		}
	}

	catalog := cr.httpc.Catalog()
	node, _, err := catalog.Node(rr.Agent, nil)
	if err != nil {
		return nil, quantum.NoAgentsFromRequest(rr)
	}

	// Quantum services are registered with an UUID for an ID for uniqueness.
	// The agent consul dictionary maps services by the ID. Since we don't know
	// UUID we're looking for, loop over all services to search by service name.
	var service *api.AgentService
	for _, srv := range node.Services {
		if srv.Service == rr.Type {
			service = srv
			break
		}
	}

	if service == nil {
		return nil, quantum.NoAgentsFromRequest(rr)
	}

	if service.Address == "" {
		service.Address = node.Node.Address
	}

	return []resolveResult{{address: fmt.Sprintf("%s:%d", service.Address, service.Port)}}, nil
}

func (cr *ClientResolver) resolveClient(results []resolveResult) (conn quantum.ClientConn, err error) {
	// TODO: Do this concurrently, first one to respond wins
	for _, result := range results {
		client := client.New(cr.config)
		conn, err = client.Dial(result.address)
		if err == nil {
			break
		}
	}
	return
}
