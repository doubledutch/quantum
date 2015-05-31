package agent

import (
	"errors"
	"strconv"
	"testing"

	"github.com/doubledutch/quantum"
	"github.com/mitchellh/multistep"
)

const (
	serverJob = "serverJob"
	testPort  = ":8814"
)

type testAgentRequest struct {
	Type int
}

type testAgentJob struct {
	*quantum.BasicJob
}

func (j *testAgentJob) Type() string {
	return serverJob
}

func (j *testAgentJob) Configure(p []byte) error {
	j.BasicJob = quantum.NewBasicJob(j)
	return nil
}

type AddStep struct{}

func (s AddStep) Run(state quantum.StateBag) error {
	var count int
	rawCount, ok := state.GetOk("count")
	if !ok {
		count = 0
	} else {
		count = (rawCount.(int)) + 1
	}

	outCh := state.Get("outCh").(chan string)
	outCh <- strconv.Itoa(count)

	state.Put("count", count)

	return nil
}

func (s AddStep) Cleanup(state quantum.StateBag) {}

type ErrorStep struct{}

func (s ErrorStep) Run(state multistep.StateBag) multistep.StepAction {
	state.Put("error", errors.New("error"))

	return multistep.ActionHalt
}

func (j *testAgentJob) Steps() []quantum.Step {
	return []quantum.Step{
		&AddStep{},
		&AddStep{},
		&AddStep{},
	}
}

func TestPort(t *testing.T) {
	p := NewPort(":8500")

	if p.Value != ":8500" {
		t.Fatal("wrong port value")
	}

	if p.Int() != 8500 {
		t.Fatal("wrong port int")
	}
}
