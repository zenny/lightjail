package util

import (
	"os"
	"os/exec"
)

type Mounter struct {
	Points []string
}

func (mounter *Mounter) runMount(mountCmd *exec.Cmd) {
	mountCmd.Stderr = os.Stdout
	if err := mountCmd.Run(); err != nil {
		panic(err)
	}
}

func (mounter *Mounter) Mount(fs, opt, from, to string) {
	mounter.runMount(exec.Command("mount", "-t", fs, "-o", opt, from, to))
	mounter.Points = append(mounter.Points, to)
}

func (mounter *Mounter) MountDev(to string) {
	mounter.runMount(exec.Command("mount", "-t", "devfs", "devfs", to))
	mounter.runMount(exec.Command("devfs", "-m", to, "rule", "-s", "4", "applyset"))
	mounter.Points = append(mounter.Points, to)
}

func (mounter *Mounter) Unmount(from string) {
	mounter.runMount(exec.Command("umount", from))
}

func (mounter *Mounter) Cleanup() {
	for i := len(mounter.Points) - 1; i >= 0; i-- {
		mounter.Unmount(mounter.Points[i])
	}
}
