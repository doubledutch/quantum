package quantum

// Registrator registers and deregisters services
type Registrator interface {
	Register(port int, reg Registry) error
	Deregister() error
}
