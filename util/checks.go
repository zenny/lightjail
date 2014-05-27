package util

import (
	"log"
	"os/exec"
)

func CheckExecutables(names ...string) {
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			log.Fatal(err)
		}
	}
}
