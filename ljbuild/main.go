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

var overrideVersion, ipAddr string

func main() {
	if _, err := exec.LookPath("ljspawn"); err != nil {
		log.Fatal(err)
	}
	if _, err := exec.LookPath("mount"); err != nil {
		log.Fatal(err)
	}
	// TODO: option to hide rootdir?
	flag.StringVar(&overrideVersion, "v", "", "Override version")
	flag.StringVar(&ipAddr, "i", "", "IPv4 address of the build jail")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("usage: ljbuild /path/to/Jailfile\n")
		os.Exit(2)
	}
	if ipAddr == "" {
		ipAddr = "192.168.1.240"
	}
	runJailfile(args[0])
}

func runJailfile(path string) {
	script := parseJailfile(readJailfile(path))
	if overrideVersion != "" {
		script.Version = overrideVersion
	}
	rootDir := getRootDir()
	worldDir := filepath.Join(rootDir, "worlds", "10.0-test") // TODO: read world version
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
	mounter := new(Mounter)
	mounter.Mount("nullfs", "ro", worldDir, mountPoint)
	mounter.Mount("unionfs", "rw", overlayDir, mountPoint)
	ljspawnCmd := exec.Command("ljspawn", "-n", filepath.Base(mountPoint), "-i", ipAddr, "-d", mountPoint, "-p", script.Buildscript)
	ljspawnCmd.Stdout = os.Stdout
	ljspawnCmd.Stderr = os.Stdout
	cleanupFn := func() {
		cleanup(mounter, mountPoint)
	}
	runner := Runner{Command: ljspawnCmd, Cleanup: cleanupFn}
	runner.Run()
	block()
}

func cleanup(mounter *Mounter, mountPoint string) {
	mounter.UnmountAll()
	syscall.Rmdir(mountPoint)
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
