package quantum

import (
	"log"
	"os"
	"testing"
)

func TestRunner(t *testing.T) {
	runner := NewBasicRunner()

	outCh := make(chan string, 1)
	sigCh := make(chan os.Signal, 1)
	go func() {
		<-outCh // info
		if out := <-outCh; out != "hello world\n" {
			t.Fatalf("'%s' != 'hello world\n'", out)
		}
	}()

	if err := runner.Run("echo hello world", outCh, sigCh); err != nil {
		t.Fatal(err)
	}
}

func TestRunnerErr(t *testing.T) {
	runner := NewBasicRunner()

	outCh := make(chan string, 1)
	sigCh := make(chan os.Signal, 1)
	go func() {
		for out := range outCh {
			log.Println(out)
		}
	}()

	if err := runner.Run("asdf", outCh, sigCh); err == nil {
		t.Fatal("runner did not exit with error")
	}
}

func TestRunnerCancel(t *testing.T) {
	runner := NewBasicRunner()

	outCh := make(chan string, 1)
	sigCh := make(chan os.Signal, 1)
	go func() {
		for _ = range outCh {
			// Consume outCh
		}
	}()

	sigCh <- os.Kill

	if err := runner.Run("sleep 1", outCh, sigCh); err == nil {
		t.Fatal("runner did not exit with error")
	}
}
