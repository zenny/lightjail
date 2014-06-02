package main

import (
	"encoding/json"
	"fmt"
	"github.com/myfreeweb/lightjail/util"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Overlay struct {
	Name    string
	Version string
	From    *Overlay
}

func (overlay *Overlay) Save(path string) {
	bytes, err := json.Marshal(overlay)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path, bytes, 0444)
	if err != nil {
		panic(err)
	}
}

func ReadOverlay(path string) Overlay {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var overlay Overlay
	err = json.Unmarshal(bytes, &overlay)
	if err != nil {
		panic(err)
	}
	return overlay
}

func (overlay *Overlay) GetFromPaths(rootDir string) []string {
	paths := []string{}
	if overlay.From != nil {
		unversionedPath := filepath.Join(rootDir, overlay.From.Name)
		path := filepath.Join(unversionedPath, util.FindMaxVersionFile(unversionedPath, overlay.From.Version))
		parent := ReadOverlay(filepath.Join(path, "overlay.json"))
		paths = append(paths, parent.GetFromPaths(rootDir)...)
		paths = append(paths, path)
	}
	return paths
}

type Script struct {
	Overlay
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
