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

type RunnerCommand struct {
	Command *exec.Cmd
	Done    chan int
}

type Runner struct {
	Commands []RunnerCommand
}

func (runner *Runner) Run(cmd *exec.Cmd) chan int {
	done := make(chan int, 1)
	runner.Commands = append(runner.Commands, RunnerCommand{cmd, done})
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	go func() {
		state, _ := cmd.Process.Wait()
		done <- int(state.Sys().(syscall.WaitStatus))
	}()
	return done
}

func (runner *Runner) handleInterrupts() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-interrupt
		var waitGrp sync.WaitGroup
		waitGrp.Add(len(runner.Commands))
		for _, rcmd := range runner.Commands {
			go func() {
				cmd := rcmd.Command
				cmd.Process.Signal(syscall.SIGINT) // INT ljspawn -> it will TERM the process it runs
				select {
				case <-rcmd.Done:
				case <-time.After(2 * time.Second):
					cmd.Process.Signal(syscall.SIGTERM) // TERM ljspawn -> it will KILL the process it runs
				case <-time.After(8 * time.Second):
					cmd.Process.Kill() // Something is fucked up and we have to kill ljspawn
				}
				waitGrp.Done()
			}()
		}
		waitGrp.Wait()
	}()
}
