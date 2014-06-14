package main

import (
	"github.com/myfreeweb/lightjail/util"
	"path/filepath"
)

type Process struct {
	Name         string
	WorldVersion string
	Overlay      util.Overlay
	MemLimit     int
	Script       string
	Up           bool
	NeedsToBeUp  bool

	mountPoint string
}

func (process *Process) WorldDir() string {
	return filepath.Join(util.RootDir(), "worlds", process.WorldVersion)
}

func (process *Process) OverlayDir() string {
	return filepath.Join(util.RootDir(), process.Overlay.Name, process.Overlay.Version)
}
