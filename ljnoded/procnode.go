package main

import (
	"github.com/myfreeweb/gomaplog"
	"github.com/myfreeweb/lightjail/util"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
)

type ProcNode struct {
	sync.Mutex
	Processes map[string]*Process
	Runner    *util.Runner
	Mounter   *util.Mounter
}

func NewProcNode() *ProcNode {
	node := new(ProcNode)
	node.Processes = make(map[string]*Process)
	node.Runner = new(util.Runner)
	node.Mounter = new(util.Mounter)
	return node
}

func (node *ProcNode) StartHandlingInterrupts() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		i := <-interrupt
		logger.Notice("Interrupted by a signal", gomaplog.Extras{"signal": i})
		node.Runner.CleanupWithHook(func(uid string) {
			proc := node.Processes[uid]
			proc.Up = false
			writeProcess(uid, proc)
		})
		node.Mounter.Cleanup()
		os.Exit(1)
	}()
}

func (node *ProcNode) HandleProcesses(processes map[string]*Process) {
	for uid, newProc := range processes {
		oldProc, exists := node.Processes[uid]
		logData := gomaplog.Extras{"uid": uid, "name": newProc.Name, "world": newProc.WorldVersion, "overlay": newProc.OverlayDir(), "mem_limit": newProc.MemLimit}
		if exists {
			if newProc.NeedsToBeUp { // Restart
				if newProc.Name != oldProc.Name { // TODO: actual comparison
					logger.Info("Restarting process", logData)
					node.Stop(uid)
					node.Start(newProc, uid)
					node.SetProcess(newProc, uid)
				}
			} else if !newProc.NeedsToBeUp { // Stop
				logger.Info("Stopping process", logData)
				node.Stop(uid)
			} else {
				logger.Info("Nothing to do", logData)
			}
		} else {
			if newProc.NeedsToBeUp { // Start
				logger.Info("Starting process", logData)
				node.Start(newProc, uid)
				node.SetProcess(newProc, uid)
			}
		}
	}
}

func (node *ProcNode) Start(process *Process, uid string) {
	ipAddr := util.FreeIPAddr()
	jsonOption := ""
	if options.logJson {
		jsonOption = "-j"
	}
	process.mountPoint = util.MustTempDir("ljnoded-")
	mount(process, node.Mounter)
	cmd := exec.Command("ljspawn", "-n", process.Name, "-m", strconv.Itoa(process.MemLimit), "-i", ipAddr, "-f", options.ipIface, "-d", process.mountPoint, "-p", process.Script, jsonOption)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	procDone := node.Runner.Run(cmd, uid)
	go func() {
		<-procDone
		unmount(process, node.Mounter)
		logData := gomaplog.Extras{"uid": uid, "name": process.Name, "world": process.WorldVersion, "overlay": process.OverlayDir(), "mem_limit": process.MemLimit}
		logger.Info("Process finished", logData)
		// TODO: optional auto restart if NeedsToBeUp
	}()
}

func (node *ProcNode) Stop(uid string) {
	node.Processes[uid].Up = false
	node.Runner.StopByUid(uid)
	delete(node.Processes, uid)
}

func mount(process *Process, mounter *util.Mounter) {
	mounter.Mount("nullfs", "ro", process.WorldDir(), process.mountPoint)
	for _, path := range process.Overlay.GetFromPaths() {
		mounter.Mount("unionfs", "ro", path, process.mountPoint)
	}
	mounter.Mount("unionfs", "rw", process.OverlayDir(), process.mountPoint) // TODO: ro here + ephemeral dir
	mounter.MountDev(filepath.Join(process.mountPoint, "dev"))
}

func unmount(process *Process, mounter *util.Mounter) {
	mounter.Unmount(filepath.Join(process.mountPoint, "dev"))
	for i := 0; i < len(process.Overlay.GetFromPaths())+2; i++ {
		mounter.Unmount(process.mountPoint)
	}
}

func (node *ProcNode) SetProcess(process *Process, uid string) {
	node.Lock()
	node.Processes[uid] = process
	node.Unlock()
}
