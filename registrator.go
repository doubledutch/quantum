package quantum

import "github.com/hashicorp/go-multierror"

// Registrator registers and deregisters services
type Registrator interface {
	Register(port int, reg Registry) error
	Deregister() error
}

// MultiRegistrator holds multiple Registry instances
type MultiRegistrator struct {
	Registrators []Registrator
}

// Register calls Register on Registries
func (r *MultiRegistrator) Register(port int, reg Registry) error {
	var result error

	for _, registrator := range r.Registrators {
		if err := registrator.Register(port, reg); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

// Deregister calls Deregister on Registries
func (r *MultiRegistrator) Deregister() error {
	var result error

	for _, registrator := range r.Registrators {
		if err := registrator.Deregister(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}
