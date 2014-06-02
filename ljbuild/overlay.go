package main

import (
	"encoding/json"
	"github.com/myfreeweb/lightjail/util"
	"io/ioutil"
	"path/filepath"
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
