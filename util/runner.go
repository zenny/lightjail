package util

import (
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type RunnerCommand struct {
	Command *exec.Cmd
	Done    chan int
}

func (rcmd *RunnerCommand) Stop() {
	proc := rcmd.Command.Process
	proc.Signal(syscall.SIGINT) // INT ljspawn -> it will TERM the process it runs
	select {
	case <-rcmd.Done:
	case <-time.After(2 * time.Second):
		proc.Signal(syscall.SIGTERM) // TERM ljspawn -> it will KILL the process it runs
	case <-time.After(8 * time.Second):
		proc.Kill() // Something is fucked up and we have to KILL ljspawn
	}
}

type Runner struct {
	sync.Mutex
	Commands []RunnerCommand
}

func (runner *Runner) Run(cmd *exec.Cmd) chan int {
	done := make(chan int, 1)
	runner.Lock()
	idx := len(runner.Commands) // Clever hack: the length is equal to the index of what we're going to append
	runner.Commands = append(runner.Commands, RunnerCommand{cmd, done})
	runner.Unlock()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	go func() {
		state, _ := cmd.Process.Wait()
		runner.Lock()
		runner.Commands = append(runner.Commands[:idx], runner.Commands[idx+1:]...)
		runner.Unlock()
		done <- int(state.Sys().(syscall.WaitStatus))
		close(done)
	}()
	return done
}

func (runner *Runner) Stop() {
	var waitGrp sync.WaitGroup
	waitGrp.Add(len(runner.Commands))
	for _, rcmd := range runner.Commands {
		go func() {
			rcmd.Stop()
			waitGrp.Done()
		}()
	}
	waitGrp.Wait()
}
