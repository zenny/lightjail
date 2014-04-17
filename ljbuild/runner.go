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
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
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
		sig := <-interrupt
		runner.Command.Process.Signal(sig)
		<-time.After(2 * time.Second)
		done <- runner.Command.Process.Kill()
	}()
}

func block() {
	for {
		time.Sleep(1 * time.Second)
	}
}
