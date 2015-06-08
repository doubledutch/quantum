package quantum

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/doubledutch/mux"
)

const (
	// RequestType is a mux type for requests
	RequestType = uint8(67)
)

var (
	// ErrSigReceived is used when a signal is received while running a job
	ErrSigReceived = errors.New("Signal Received")
)

// Request contains Type and Data. Type is the ID, Data is the request data
type Request struct {
	Type string
	Data []byte
}

// NewRequest creates a request using the request type and request data as strings
func NewRequest(rt, rd string) Request {
	return Request{
		Type: rt,
		Data: []byte(rd),
	}
}

// Routable defines an interface for routing an object
type Routable interface {
	Type() string
}

// RoutableRequest creates a Request
func RoutableRequest(r Routable) Request {
	b, _ := json.Marshal(r)

	return Request{
		Type: r.Type(),
		Data: b,
	}
}

// Job is executed by Quantum
type Job interface {
	// Used to match requests to jobs
	Routable
	Configure([]byte) error
	Run(AgentConn) error
}

// StepsJob is a superset of Job, providing Steps so this job can by ran
// by BasicRun. This could be a StepJob with a BasicImplementation.
type StepsJob interface {
	Job
	Steps() []Step
}

// NewBasicJob creates a new BasicJob
func NewBasicJob(job StepsJob) *BasicJob {
	return &BasicJob{
		job: job,
	}
}

// BasicJob executes a StepsJob on Run
type BasicJob struct {
	job StepsJob
}

// Run runs the basic job
func (basic *BasicJob) Run(conn AgentConn) error {
	outCh := conn.Logs()

	// TODO: Move this to agent
	// The basic job may be able to Execute(AgentConn) error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Send response
		for output := range outCh {
			conn.Send(mux.LogType, output)
		}
		wg.Done()
	}()

	state := NewStateBag()
	state.Put("conn", conn)
	state.Put("ui", NewUI(conn))
	state.Put("runner", NewBasicRunner())

	runner := &BasicExecutor{Steps: basic.job.Steps()}
	err := runner.Run(state)

	close(outCh)
	wg.Wait()

	return err
}
