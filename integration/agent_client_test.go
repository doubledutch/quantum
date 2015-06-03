package integration

import (
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/doubledutch/lager"
	"github.com/doubledutch/mux/gob"
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

type ErrorStep struct{}

func (s ErrorStep) Run(state quantum.StateBag) error {
	return errors.New("error")
}

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
	port := ":0"

	qc := &quantum.Config{
		Pool: new(gob.Pool),
		Lager: lager.NewLogLager(&lager.LogConfig{
			Levels: lager.LevelsFromString("DIE"),
			Output: os.Stdout,
		}),
	}

	cc := &quantum.ConnConfig{
		Timeout: 100 * time.Millisecond,
		Config:  qc,
	}

	agent := agent.New(&agent.Config{
		Port:       port,
		ConnConfig: cc,
	})
	agent.Add(new(testAgentJob))

	l, agentAddr := listenTCP()
	go func() {
		if err := agent.Accept(l); err != nil {
			t.Fatal(err)
			t.FailNow()
		}
	}()

	request := quantum.NewRequest(serverJob, "{}")
	client := client.New(cc)

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

/*
func TestClientAgentServer(t *testing.T) {
	server := NewServer()
	ls, serverAddr := listenTCP()
	go func() {
		// Register agent record
		if err := server.Accept(ls); err != nil {
			t.Fatal(err)
			t.FailNow()
		}
		// ClientResolve
		if err := server.Accept(ls); err != nil {
			t.Fatal(err)
			t.FailNow()
		}
	}()

	la, agentAddr := listenTCP()
	parts := strings.Split(agentAddr, ":")
	agent := NewAgent(":" + parts[1])
	agent.Register(&testAgentJob{})
	go func() {
		if err := agent.Accept(la); err != nil {
			t.Fatal(err)
		}
	}()

	if err := agent.Announce(serverAddr); err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	cr := NewClientResolver()
	configs, err := cr.Resolve(serverAddr, "test", serverJob, "{}")
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	outCh := make(chan string)
	sigCh := make(chan os.Signal)
	defer close(sigCh)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for _ = range outCh {
			// Consume the channel
		}
		wg.Done()
	}()

	if err := cr.RunWith(configs, outCh, sigCh); err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	close(outCh)
	wg.Wait()
}*/
