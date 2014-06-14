package main

import (
	"fmt"
	"github.com/myfreeweb/lightjail/util"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type Buildjail struct {
	Script     *Script
	MemLimit   int
	MountPoint string
	RootDir    string
	Name       string
}

func (jail *Buildjail) GetWorldDir() string {
	return filepath.Join(jail.RootDir, "worlds", jail.Script.WorldVersion)
}

func (jail *Buildjail) GetFromPaths() []string {
	return jail.Script.Overlay.GetFromPaths()
}

func (jail *Buildjail) MustMakeOverlayDir() {
	if err := os.MkdirAll(jail.GetOverlayPath(), 0774); err != nil {
		panic(err)
	}
}

func (jail *Buildjail) CopyFiles() {
	if jail.Script.CopyDst != "" {
		if err := exec.Command("cp", "-R", filepath.Dir(jail.Script.ScriptPath)+"/", filepath.Join(jail.GetOverlayPath(), jail.Script.CopyDst)).Run(); err != nil {
			panic(err)
		}
	}
}
func (jail *Buildjail) GetOverlayPath() string {
	return filepath.Join(jail.RootDir, jail.Script.Name, jail.Script.Version)
}

func (jail *Buildjail) MustValidate() {
	jail.Script.MustValidate()
	if _, err := os.Stat(jail.GetWorldDir()); err != nil {
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("World does not exist: %s", jail.Script.WorldVersion))
		} else {
			panic(fmt.Sprintf("Cannot stat world: %s -- probably the server is not set up correctly", jail.Script.WorldVersion))
		}
	}
	if _, err := os.Stat(jail.GetOverlayPath()); err == nil {
		panic(fmt.Sprintf("Directory already exists: %s", jail.GetOverlayPath()))
	}
}

func (jail *Buildjail) Cmd() *exec.Cmd {
	jsonOption := ""
	if options.logJson {
		jsonOption = "-j"
	}
	ljspawnCmd := exec.Command("ljspawn", "-n", jail.Name, "-m", strconv.Itoa(jail.MemLimit), "-i", options.ipAddr, "-f", options.ipIface, "-d", jail.MountPoint, "-p", jail.Script.Buildscript, jsonOption)
	ljspawnCmd.Stdout = os.Stdout
	ljspawnCmd.Stderr = os.Stdout
	return ljspawnCmd
}

func (jail *Buildjail) Mount(mounter *util.Mounter) {
	mounter.Mount("nullfs", "ro", jail.GetWorldDir(), jail.MountPoint)
	for _, path := range jail.GetFromPaths() {
		mounter.Mount("unionfs", "ro", path, jail.MountPoint)
	}
	mounter.Mount("unionfs", "rw", jail.GetOverlayPath(), jail.MountPoint)
	mounter.MountDev(filepath.Join(jail.MountPoint, "dev"))
}
