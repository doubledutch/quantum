package quantum

// Registry adds jobs, provides access to jobs by type, and all job types.
type Registry interface {
	Add(job Job)
	Get(request Request) (Job, error)
	Types() (types []string)
}
