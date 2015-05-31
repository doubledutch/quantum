package inmemory

import (
	"errors"
	"log"

	"github.com/doubledutch/quantum"
)

var (
	// ErrJobNotFound = Register.Get requests job is not found within Registery
	ErrJobNotFound = errors.New("Job Not Found")
)

// NewRegistry creates a Router, initializing the job map.
func NewRegistry() quantum.Registry {
	return &Registry{
		jobs: make(map[string]quantum.Job),
	}
}

// Registry stores jobs in a map
type Registry struct {
	jobs map[string]quantum.Job
}

// Add registers a job with the Registry
func (r *Registry) Add(job quantum.Job) {
	r.jobs[job.Type()] = job
}

// Get returns the Job corresponding to the Request.
// If no such job exists, an error is returned.
func (r *Registry) Get(request quantum.Request) (quantum.Job, error) {
	log.Printf("attempting job with type: %v\n", request.Type)
	job, ok := r.jobs[request.Type]
	if !ok {
		log.Printf("job not found with type: %v", request.Type)
		return nil, ErrJobNotFound
	}

	err := job.Configure(request.Data)
	if err != nil {
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
