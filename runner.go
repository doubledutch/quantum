package quantum

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"

	"github.com/mitchellh/iochan"
)

// Runner defines an interface for running commands
type Runner interface {
	Run(string, chan<- string, <-chan os.Signal) error
}

// NewBasicRunner returns a basic runner
func NewBasicRunner() (r Runner) {
	return &BasicRunner{}
}

// BasicRunner is a basic implementation of Runner
type BasicRunner struct{}

// Run runs a command and captures the output of the command, while listening
// for and sending signals to the process.
func (r *BasicRunner) Run(cmd string,
	outCh chan<- string,
	sigCh <-chan os.Signal) error {
	var shell, flag string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	} else {
		envShell := os.Getenv("SHELL")
		if envShell != "" {
			shell = envShell
		} else {
			shell = "/bin/bash"
		}
		flag = "-c"
	}
	// Tell the client what we're running.
	// Note: the tests expect this
	outCh <- "Running " + cmd + "\n"
	ec := exec.Command(shell, flag, cmd)
	return run(ec, outCh, sigCh)
}

func run(cmd *exec.Cmd, outCh chan<- string, sigCh <-chan os.Signal) error {
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Create the channels we'll use for data
	exitCh := make(chan int, 1)
	doneCh := make(chan interface{}, 1)
	stdoutCh := iochan.DelimReader(outPipe, '\n')
	stderrCh := iochan.DelimReader(errPipe, '\n')
	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		select {
		case <-doneCh:
			return
		case sig := <-sigCh:
			cmd.Process.Signal(sig)
		}
	}()

	// Start the goroutine to watch for the exit
	go func() {
		exitStatus := 0

		err := cmd.Wait()
		doneCh <- struct{}{}

		if exitErr, ok := err.(*exec.ExitError); ok {
			exitStatus = 1

			// There is no process-independent way to get the REAL
			// exit status so we just try to go deeper.
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
			}
		}

		exitCh <- exitStatus
	}()

	var streamWg sync.WaitGroup
	streamWg.Add(2)

	streamFunc := func(ch <-chan string) {
		defer streamWg.Done()
		for data := range ch {
			if data != "" {
				outCh <- data
			}
		}
	}

	go streamFunc(stdoutCh)
	go streamFunc(stderrCh)

	exitStatus := <-exitCh
	streamWg.Wait()

	if exitStatus != 0 {
		return fmt.Errorf("run failed with exit code: %v", exitStatus)
	}
	return nil
}
