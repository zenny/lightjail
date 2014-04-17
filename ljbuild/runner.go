package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type Runner struct {
	Command *exec.Cmd
	Cleanup func()
}

func (runner *Runner) Run() {
	done := make(chan error, 1)
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	signal.Notify(interrupts, syscall.SIGTERM)
	if err := runner.Command.Start(); err != nil {
		log.Fatal(err)
	}
	go func() {
		done <- runner.Command.Wait()
	}()
	go func() {
		err := <-done
		runner.Cleanup()
		if err == nil {
			os.Exit(0)
		} else {
			log.Fatal(err)
		}
	}()
	go func() {
		<-interrupts
		runner.Command.Process.Signal(os.Interrupt)
		select {
		case <-time.After(2 * time.Second):
			runner.Command.Process.Kill()
			<-done
		}
	}()
}

func block() {
	for {
		time.Sleep(1 * time.Second)
	}
}
