package main

import (
	"fmt"
	"github.com/myfreeweb/lightjail/util"
	"strings"
)

type Script struct {
	util.Overlay
	WorldVersion string
	Buildscript  string
	CopyDst      string
	ScriptPath   string
}

func (script *Script) MustValidate() {
	errors := []string{}

	if script.Name == "" {
		errors = append(errors, "A 'name' directive must be present")
	}

	if !strings.Contains(script.Name, "/") {
		errors = append(errors, "Container name must contain at least one subdirectory (slash), i.e. me/myapp, not just myapp")
	}

	if strings.Contains(script.Name, "..") {
		errors = append(errors, "Container name must not contain '..'")
	}

	if strings.Contains(script.Name, ":") {
		errors = append(errors, "Container name must not contain ':'") // because rctl
	}

	if strings.Contains(script.Version, "..") {
		errors = append(errors, "Version must not contain '..'")
	}

	if strings.Contains(script.Version, "/") {
		errors = append(errors, "Version must not contain '/'")
	}

	if strings.Contains(script.CopyDst, "..") {
		errors = append(errors, "Copy must not contain '..'")
	}

	if strings.Contains(script.From.Name, "..") || strings.Contains(script.From.Version, "..") {
		errors = append(errors, "From must not contain '..'")
	}

	if strings.Contains(script.WorldVersion, "..") {
		errors = append(errors, "World must not contain '..'")
	}

	if len(errors) > 0 {
		panic(fmt.Sprintf("The Jailfile is not valid because of the following reasons:\n- %s\n", strings.Join(errors, "\n- ")))
	}
}
