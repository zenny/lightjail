package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Script struct {
	Name         string
	WorldVersion string
	Version      string
	Buildscript  string
	RootDir      string
	From         []string
}

func (script *Script) GetOverlayPath() string {
	return filepath.Join(script.RootDir, script.Name, script.Version)
}

func (script *Script) GetWorldDir() string {
	return filepath.Join(script.RootDir, "worlds", script.WorldVersion)
}

func (script *Script) Validate() {
	errors := []string{}

	if script.Name == "" {
		errors = append(errors, "A 'name' directive must be present")
	}

	if !strings.Contains(script.Name, "/") {
		errors = append(errors, "Container name must contain at least one subdirectory (slash), i.e. me/myapp, not just myapp")
	}

	if _, err := os.Stat(script.GetWorldDir()); err != nil {
		if os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("World does not exist: %s", script.WorldVersion))
		} else {
			errors = append(errors, fmt.Sprintf("Cannot stat world: %s -- probably the server is not set up correctly", script.WorldVersion))
		}
	}

	if _, err := os.Stat(script.GetOverlayPath()); err == nil {
		errors = append(errors, fmt.Sprintf("Directory already exists: %s", script.GetOverlayPath()))
	}

	if len(script.From) > 1 {
		errors = append(errors, "Only one 'from' directive can be present")
	}

	if len(errors) > 0 {
		log.Fatalf("The Jailfile is not valid because of the following reasons:\n- %s\n", strings.Join(errors, "\n- "))
	}
}
