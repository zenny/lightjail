package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Runner struct {
	Command *exec.Cmd
	Done    chan int
}

func (runner *Runner) Run() chan int {
	if err := runner.Command.Start(); err != nil {
		log.Fatal(err)
	}
	go func() {
		state, _ := runner.Command.Process.Wait()
		runner.Done <- int(state.Sys().(syscall.WaitStatus))
	}()
	runner.handleInterrupts()
	return runner.Done
}

func (runner *Runner) handleInterrupts() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-interrupt
		// for _, cmd := range runner.Commands {
		cmd := runner.Command
		var waitGrp sync.WaitGroup
		waitGrp.Add(1)
		go func() {
			cmd.Process.Signal(syscall.SIGINT) // INT ljspawn -> it will TERM the process it runs
			select {
			case <-runner.Done:
			case <-time.After(2 * time.Second):
				cmd.Process.Signal(syscall.SIGTERM) // TERM ljspawn -> it will KILL the process it runs
			case <-time.After(8 * time.Second):
				cmd.Process.Kill() // Something is fucked up and we have to kill ljspawn
			}
			waitGrp.Done()
		}()
		waitGrp.Wait()
		// }
	}()
}
