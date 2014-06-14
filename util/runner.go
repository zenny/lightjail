package util

import (
	"errors"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var CmdNotFound = errors.New("Command not found by uid")

type RunnerCommand struct {
	Command *exec.Cmd
	Done    chan int
	Uid     string
}

func (rcmd *RunnerCommand) Stop() error {
	proc := rcmd.Command.Process
	proc.Signal(syscall.SIGINT) // INT ljspawn -> it will TERM the process it runs
	select {
	case <-rcmd.Done:
	case <-time.After(2 * time.Second):
		proc.Signal(syscall.SIGTERM) // TERM ljspawn -> it will KILL the process it runs
		select {
		case <-rcmd.Done:
		case <-time.After(6 * time.Second):
			proc.Kill() // Something is fucked up and we have to KILL ljspawn
		}
	}
	return nil
}

type Runner struct {
	sync.Mutex
	Commands []RunnerCommand
}

func (runner *Runner) Run(cmd *exec.Cmd, uid string) chan int {
	done := make(chan int, 1)
	runner.Lock()
	defer runner.Unlock()
	idx := len(runner.Commands) // Clever hack: the length is equal to the index of what we're going to append
	runner.Commands = append(runner.Commands, RunnerCommand{cmd, done, uid})
	if err := cmd.Start(); err != nil {
		panic(err)
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

func (runner *Runner) StopByUid(uid string) error {
	runner.Lock()
	defer runner.Unlock()
	empty := RunnerCommand{}
	rcmd := &empty
	for _, v := range runner.Commands {
		if v.Uid == uid {
			rcmd = &v
		}
	}
	if rcmd == &empty {
		return CmdNotFound
	}
	return rcmd.Stop()
}

func (runner *Runner) CleanupWithHook(uidHook func(string)) {
	var waitGrp sync.WaitGroup
	waitGrp.Add(len(runner.Commands))
	for _, rcmd := range runner.Commands {
		go func() {
			rcmd.Stop()
			uidHook(rcmd.Uid)
			waitGrp.Done()
		}()
	}
	waitGrp.Wait()
}

func (runner *Runner) Cleanup() {
	runner.CleanupWithHook(func(uid string) {})
}
