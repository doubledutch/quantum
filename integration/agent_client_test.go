package integration

import (
	"log"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/doubledutch/quantum"
	"github.com/doubledutch/quantum/agent"
	"github.com/doubledutch/quantum/client"
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

	conn := state.Get("conn").(quantum.AgentConn)
	outCh := conn.Logs()
	outCh <- strconv.Itoa(count)

	state.Put("count", count)

	return nil
}

func (s AddStep) Cleanup(state quantum.StateBag) {}

func (j *testAgentJob) Steps() []quantum.Step {
	return []quantum.Step{
		&AddStep{},
		&AddStep{},
		&AddStep{},
	}
}

func listenTCP() (net.Listener, string) {
	l, err := net.Listen("tcp", "127.0.0.1:0") // any avaiable address
	if err != nil {
		log.Fatalf("net tcp listen :0 %v", err)
	}
	return l, l.Addr().String()
}

func TestClientAgent(t *testing.T) {
	addr := ":0"

	agent := agent.New(&agent.Config{
		Addr: addr,
	})
	agent.Add(new(testAgentJob))

	l, agentAddr := listenTCP()
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.Serve(conn); err != nil {
			t.Fatal(err)
		}
	}()

	request := quantum.NewRequest(serverJob, "{}")
	client := client.New(nil)

	conn, err := client.Dial(agentAddr)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for _ = range conn.Logs() {
			// Consume the channel
		}
		wg.Done()
	}()

	if err := conn.Run(request); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
