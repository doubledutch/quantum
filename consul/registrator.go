package consul

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/doubledutch/quantum"
	"github.com/doubledutch/lager"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-multierror"
)

// NewRegistrator creates a Registrator
func NewRegistrator(address string, lgr lager.Lager) quantum.Registrator {
	return &Registrator{
		address: address,
		lgr:     lgr,
	}
}

// Registrator uses consul to implement quantum.Registrator
type Registrator struct {
	serviceIDs []string
	address    string

	lgr lager.Lager
}

// Register will register types with Consul
func (r *Registrator) Register(port int, reg quantum.Registry) error {
	client, err := api.NewClient(&api.Config{
		Address: r.address,
	})
	if err != nil {
		r.lgr.Errorf("Unable to connect to Consul: %s\n", err)
		return err
	}

	merr := &multierror.Error{}

	// Relies on local consul agent
	agent := client.Agent()
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
	client, err := api.NewClient(&api.Config{
		Address: r.address,
	})
	if err != nil {
		r.lgr.Errorf("[ERROR] Unable to connect to Consul: %s\n", err)
		return err
	}
	agent := client.Agent()

	for _, serviceID := range r.serviceIDs {
		if err := agent.ServiceDeregister(serviceID); err != nil {
			multierror.Append(merr, err)
		}
	}

	return merr.ErrorOrNil()
}
