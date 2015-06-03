package inmemory

import (
	"strconv"

	"github.com/doubledutch/quantum"
)

// Registrator registers jobs locally
type Registrator struct {
	// Type -> Address
	Jobs map[string]string
}

// NewRegistrator creates a new Registrator
func NewRegistrator() *Registrator {
	return &Registrator{
		Jobs: make(map[string]string),
	}
}

// Register will register a Registry locally
func (r *Registrator) Register(port int, reg quantum.Registry) error {
	for _, t := range reg.Types() {
		r.Jobs[t] = "0.0.0.0:" + strconv.Itoa(port)
	}

	return nil
}

// Deregister will register the jobs locally
func (r *Registrator) Deregister() error {
	r.Jobs = nil
	return nil
}
