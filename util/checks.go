package util

import (
	"os/exec"
)

func MustHaveExecutables(names ...string) {
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			panic(err)
		}
	}
}
