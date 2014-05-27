package util

import (
	"fmt"
	"os/exec"
)

type Rctl struct {
	Limits []string
}

func (rctl *Rctl) AddLimit(lim string) {
	rctl.Limits = append(rctl.Limits, lim)
	if err := exec.Command("rctl", "-a", lim).Run(); err != nil {
		panic(err)
	}
}

func (rctl *Rctl) LimitJailRam(jname, soft, hard string) {
	rctl.AddLimit(fmt.Sprintf("jail:%s:vmemoryuse:sighup=%s", jname, soft))
	rctl.AddLimit(fmt.Sprintf("jail:%s:vmemoryuse:deny=%s", jname, hard))
}

func (rctl *Rctl) Cleanup() {
	for _, lim := range rctl.Limits {
		if err := exec.Command("rctl", "-r", lim).Run(); err != nil {
			panic(err)
		}
	}
}
