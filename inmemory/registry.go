package inmemory

import (
	"errors"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/quantum"
)

var (
	// ErrJobNotFound = Register.Get requests job is not found within Registery
	ErrJobNotFound = errors.New("Job Not Found")
)

// NewRegistry creates a Router, initializing the job map.
func NewRegistry(lgr lager.Lager) quantum.Registry {
	return &Registry{
		lgr:  lgr,
		jobs: make(map[string]quantum.Job),
	}
}

// Registry stores jobs in a map
type Registry struct {
	lgr lager.Lager

	jobs map[string]quantum.Job
}

// Add registers a job with the Registry
func (r *Registry) Add(job quantum.Job) {
	r.lgr.Debugf("Adding job: %s\n", job)
	r.jobs[job.Type()] = job
}

// Get returns the Job corresponding to the Request.
// If no such job exists, an error is returned.
func (r *Registry) Get(request quantum.Request) (quantum.Job, error) {
	r.lgr.Infof("attempting job with type: %v\n", request.Type)
	job, ok := r.jobs[request.Type]
	if !ok {
		r.lgr.Errorf("job not found with type: %v", request.Type)
		return nil, ErrJobNotFound
	}

	err := job.Configure(request.Data)
	if err != nil {
		r.lgr.Errorf("job configure error: type: %s, data: %s", request.Type, request.Data)
		return nil, err
	}

	return job, nil
}

// Types returns job types for Registry
func (r *Registry) Types() (types []string) {
	for key := range r.jobs {
		types = append(types, key)
	}
	return
}
