package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Script struct {
	WorldVersion string
	Type         string
	Path         string
	Version      string
	Buildscript  string
}

func readJailfile(path string) string {
	fileSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(fileSrc)
}

func getRootDir() string {
	rootDir := os.Getenv("LIGHTJAIL_ROOT")
	if rootDir == "" {
		rootDir = "/usr/local/lj"
	}
	return rootDir
}

func (script *Script) getOverlayPath(rootDir string) string {
	return filepath.Join(rootDir, script.Type+"s", script.Path, script.Version)
}

func mount(fs, opt, from, to string) {
	mountCmd := exec.Command("mount", "-t", fs, "-o", opt, from, to)
	mountCmd.Stderr = os.Stdout
	if err := mountCmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func unmount(from string) {
	unmountCmd := exec.Command("umount", from)
	unmountCmd.Stderr = os.Stdout
	if err := unmountCmd.Run(); err != nil {
		log.Print(err)
	}
}

func main() {
	if _, err := exec.LookPath("ljspawn"); err != nil {
		log.Fatal(err)
	}
	if _, err := exec.LookPath("mount"); err != nil {
		log.Fatal(err)
	}
	// TODO: option to hide rootdir?
	var overrideVersion string
	flag.StringVar(&overrideVersion, "v", "", "Override version")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("usage: ljbuild /path/to/Jailfile\n")
		os.Exit(2)
	}
	script := parseJailfile(readJailfile(args[0]))
	if overrideVersion != "" {
		script.Version = overrideVersion
	}
	rootDir := getRootDir()
	worldDir := filepath.Join(rootDir, "worlds", "10.0-RELEASE") // TODO: read world version
	overlayDir := script.getOverlayPath(rootDir)
	if _, err := os.Stat(overlayDir); err != nil {
		if os.IsExist(err) {
			log.Fatalf("Directory already exists: %s\n", overlayDir)
		}
	}
	if err := os.MkdirAll(overlayDir, 0644); err != nil {
		log.Fatal(err)
	}
	mountPoint, err := ioutil.TempDir("", "ljbuild-")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Building %s %s version %s\n", script.Type, script.Path, script.Version)
	mount("nullfs", "ro", worldDir, mountPoint)
	mount("unionfs", "rw", overlayDir, mountPoint)
	ljspawnCmd := exec.Command("ljspawn", "-n", filepath.Base(mountPoint), "-d", mountPoint, "-p", script.Buildscript)
	ljspawnCmd.Stdout = os.Stdout
	ljspawnCmd.Stderr = os.Stdout
	if err := ljspawnCmd.Run(); err != nil {
		log.Print(err)
	}
	unmount(mountPoint)
	unmount(mountPoint)
	syscall.Rmdir(mountPoint)
}