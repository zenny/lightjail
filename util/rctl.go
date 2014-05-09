package util

import (
	"fmt"
	"log"
	"os/exec"
)

type Rctl struct {
	Limits []string
}

func (rctl *Rctl) LimitJailRam(jname, soft, hard string) {
	softLim := fmt.Sprintf("jail:%s:vmemoryuse:sighup=%s", jname, soft)
	rctl.Limits = append(rctl.Limits, softLim)
	if err := exec.Command("rctl", "-a", softLim).Run(); err != nil {
		log.Fatal(err)
	}
	hardLim := fmt.Sprintf("jail:%s:vmemoryuse:deny=%s", jname, hard)
	rctl.Limits = append(rctl.Limits, hardLim)
	if err := exec.Command("rctl", "-a", hardLim).Run(); err != nil {
		log.Fatal(err)
	}
}

func (rctl *Rctl) RemoveAll() {
	for _, lim := range rctl.Limits {
		if err := exec.Command("rctl", "-r", lim).Run(); err != nil {
			log.Fatal(err)
		}
	}
}
