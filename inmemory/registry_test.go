package inmemory

import (
	"errors"
	"testing"

	"github.com/doubledutch/quantum"
)

const (
	registryJob = "registryJob"
)

type testRegistryJob struct{}

func (j *testRegistryJob) Type() string {
	return registryJob
}

func (j *testRegistryJob) Configure(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	if p[0] == 1 {
		return errors.New("this is an error")
	}
	return nil
}

func (j *testRegistryJob) Run(conn quantum.AgentConn) error {
	return nil
}

func TestAgentRouterGet(t *testing.T) {
	r := NewRegistry()

	job := &testRegistryJob{}
	r.Add(job)

	request := quantum.Request{
		Type: registryJob,
		Data: []byte{},
	}
	actual, err := r.Get(request)
	if err != nil {
		t.Fatal(err)
	}
	if actual.Type() != job.Type() {
		t.Fatal("error routing jobs")
	}
}

func TestAgentRouterGetErr(t *testing.T) {
	r := NewRegistry()

	request := quantum.Request{
		Type: registryJob,
		Data: []byte{},
	}

	if _, err := r.Get(request); err == nil {
		t.Fatal("get should fail")
	}
}

func TestNewInMemoryRegistryGetErr(t *testing.T) {
	r := NewRegistry()

	job := &testRegistryJob{}
	r.Add(job)

	b := make([]byte, 1)
	b[0] = 1
	request := quantum.Request{
		Type: registryJob,
		Data: b,
	}

	if _, err := r.Get(request); err.Error() != "this is an error" {
		t.Fatal(err)
	}
}
