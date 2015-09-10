package consul

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/doubledutch/lager"
	"github.com/doubledutch/quantum"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-multierror"
)

// NewRegistrator creates a Registrator
func NewRegistrator(httpAddr string, lgr lager.Lager) quantum.Registrator {
	return &Registrator{
		httpAddr: httpAddr,
		lgr:      lgr,
	}
}

// Registrator uses consul to implement quantum.Registrator
type Registrator struct {
	serviceIDs []string
	httpAddr   string

	client *api.Client

	lgr lager.Lager
}

// Register will register types with Consul
func (r *Registrator) Register(port int, reg quantum.Registry) error {
	if r.client == nil {
		var err error
		r.client, err = api.NewClient(&api.Config{
			Address: r.httpAddr,
		})
		if err != nil {
			r.lgr.Errorf("Unable to connect to Consul: %s\n", err)
			return err
		}
	}
	merr := &multierror.Error{}

	// Relies on local consul agent
	agent := r.client.Agent()
	for _, jobType := range reg.Types() {
		ID := uuid.New()
		// We may need to set the ID ourselves to guarantee it's unique
		service := &api.AgentServiceRegistration{
			ID:   ID,
			Name: jobType,
			Port: port,
			Tags: []string{"quantum"},
		}
		if err := agent.ServiceRegister(service); err != nil {
			multierror.Append(merr, err)
		} else {
			r.serviceIDs = append(r.serviceIDs, ID)
		}
	}

	return merr.ErrorOrNil()
}

// Deregister deregisters our serviceIDs with Consul
func (r *Registrator) Deregister() error {
	merr := &multierror.Error{}
	agent := r.client.Agent()

	for _, serviceID := range r.serviceIDs {
		if err := agent.ServiceDeregister(serviceID); err != nil {
			multierror.Append(merr, err)
		}
	}

	return merr.ErrorOrNil()
}
