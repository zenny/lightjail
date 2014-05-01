package util

import (
	"log"
	"os"
	"os/exec"
)

type Mounter struct {
	Points []string
}

func (mounter *Mounter) Mount(fs, opt, from, to string) {
	mountCmd := exec.Command("mount", "-t", fs, "-o", opt, from, to)
	mountCmd.Stderr = os.Stdout
	if err := mountCmd.Run(); err != nil {
		log.Fatal(err)
	}
	mounter.Points = append(mounter.Points, to)
}

func (mounter *Mounter) Unmount(from string) {
	unmountCmd := exec.Command("umount", from)
	unmountCmd.Stderr = os.Stdout
	if err := unmountCmd.Run(); err != nil {
		log.Print(err)
	}
}

func (mounter *Mounter) UnmountAll() {
	for i := len(mounter.Points) - 1; i >= 0; i-- {
		mounter.Unmount(mounter.Points[i])
	}
}
