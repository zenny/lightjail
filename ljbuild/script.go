package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Script struct {
	WorldVersion string
	Type         string
	Path         string
	Version      string
	Buildscript  string
	RootDir      string
}

func (script *Script) GetOverlayPath() string {
	return filepath.Join(script.RootDir, script.Path, script.Version)
}

func (script *Script) GetWorldDir() string {
	return filepath.Join(script.RootDir, "worlds", script.WorldVersion)
}

func (script *Script) Validate() {
	errors := []string{}

	if !strings.Contains(script.Path, "/") {
		errors = append(errors, "Container name should contain at least one subdirectory (slash), i.e. me/myapp, not just myapp")
	}

	if _, err := os.Stat(script.GetWorldDir()); err != nil {
		if os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("World does not exist: %s", script.WorldVersion))
		}
	}

	if _, err := os.Stat(script.GetOverlayPath()); err == nil {
		errors = append(errors, fmt.Sprintf("Directory already exists: %s", script.GetOverlayPath()))
	}

	if len(errors) > 0 {
		log.Fatalf("The Jailfile is not valid because of the following reasons:\n- %s\n", strings.Join(errors, "\n- "))
	}
}
