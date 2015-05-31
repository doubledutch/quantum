package quantum

import "sync"

// Communicator defines an interface for communicating quantum connections
type Communicator interface {
	Communicate(exitCh chan struct{})
	Add(num int)
	Done()
	Wait()
}

// BasicCommunicator is a basic implementation of Communicator
type BasicCommunicator struct {
	sync.WaitGroup
	parent AgentConn
	child  ClientConn
}

// NewBasicCommunicator creates a new basic communicator between parent and child
func NewBasicCommunicator(parent AgentConn, child ClientConn) Communicator {
	bc := &BasicCommunicator{
		parent: parent,
		child:  child,
	}

	return bc
}

// Communicate faciliates communication the parent and child
func (c *BasicCommunicator) Communicate(exitCh chan struct{}) {
	c.DrainLogs()
	c.ForwardSignals(exitCh)
}

// DrainLogs drains logs from the child and sends them to the parent.
func (c *BasicCommunicator) DrainLogs() {
	c.Add(1)
	// Any logs we get from conn, send to parent conn logCh
	go func() {

	LOOP:
		for {
			select {
			case log := <-c.child.Logs():
				c.parent.Logs() <- log
			case <-c.child.IsShutdown():
				break LOOP
			case <-c.parent.IsShutdown():
				break LOOP
			}
		}
		c.Done()

	}()
}

// ForwardSignals forwards signals from the parent to the child.
// Forwarding stops when exit closes
func (c *BasicCommunicator) ForwardSignals(exit chan struct{}) {
	c.Add(1)
	go func() {
	LOOP:
		for {
			select {
			case sig := <-c.parent.Signals():
				c.child.Signals() <- sig
			case <-c.child.IsShutdown():
				break LOOP
			case <-c.parent.IsShutdown():
				break LOOP
			case <-exit:
				break LOOP
			}
		}
		c.Done()
	}()
}

// NewCommunicator creates a default communicator
func NewCommunicator(parent AgentConn, child ClientConn) Communicator {
	return NewBasicCommunicator(parent, child)
}
