package main

import (
	"encoding/json"
	"fmt"
	"github.com/myfreeweb/lightjail/util"
	"io/ioutil"
	"log"
	"os"
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
		log.Fatal(err)
	}
	err = ioutil.WriteFile(path, bytes, 0444)
	if err != nil {
		log.Fatal(err)
	}
}

func ReadOverlay(path string) Overlay {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var overlay Overlay
	err = json.Unmarshal(bytes, &overlay)
	if err != nil {
		log.Fatal(err)
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
	RootDir      string
	CopyDst      string
}

func (script *Script) GetOverlayPath() string {
	return filepath.Join(script.RootDir, script.Name, script.Version)
}

func (script *Script) GetWorldDir() string {
	return filepath.Join(script.RootDir, "worlds", script.WorldVersion)
}

func (script *Script) GetFromPaths() []string {
	return script.Overlay.GetFromPaths(script.RootDir)
}

func (script *Script) Validate() {
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

	if len(errors) > 0 {
		log.Fatalf("The Jailfile is not valid because of the following reasons:\n- %s\n", strings.Join(errors, "\n- "))
	}
}
