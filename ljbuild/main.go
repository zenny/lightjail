package main

import (
	"flag"
	"github.com/myfreeweb/lightjail/util"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type Script struct {
	WorldVersion string
	Type         string
	Path         string
	Version      string
	Buildscript  string
}

var overrideVersion, ipAddr, ipIface string

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
	flag.StringVar(&ipIface, "f", "", "network interface for the build jail")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}
	if ipAddr == "" {
		ipAddr = "192.168.1.240"
	}
	if ipIface == "" {
		ipIface = "lo0"
	}
	runJailfile(args[0])
}

func runJailfile(path string) {
	script := parseJailfile(readJailfile(path), overrideVersion)
	rootDir := getRootDir()
	worldDir := filepath.Join(rootDir, "worlds", script.WorldVersion)
	if _, err := os.Stat(worldDir); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("World does not exist: %s\n", script.WorldVersion)
		}
	}
	overlayDir := script.getOverlayPath(rootDir)
	if _, err := os.Stat(overlayDir); err == nil {
		log.Fatalf("Directory already exists: %s\n", overlayDir)
	}
	if err := os.MkdirAll(overlayDir, 0774); err != nil {
		log.Fatal(err)
	}
	mountPoint, err := ioutil.TempDir("", "ljbuild-")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Building %s version %s\n", script.Path, script.Version)
	mounter := new(util.Mounter)
	mounter.Mount("nullfs", "ro", worldDir, mountPoint)
	mounter.Mount("unionfs", "rw", overlayDir, mountPoint)
	ljspawnCmd := exec.Command("ljspawn", "-n", filepath.Base(mountPoint), "-i", ipAddr, "-f", ipIface, "-d", mountPoint, "-p", script.Buildscript)
	ljspawnCmd.Stdout = os.Stdout
	ljspawnCmd.Stderr = os.Stdout
	runner := new(util.Runner)
	runner.HandleInterrupts()
	code := <-runner.Run(ljspawnCmd)
	time.Sleep(300 * time.Millisecond) // Wait for jail removal, just in case
	mounter.UnmountAll()
	syscall.Rmdir(mountPoint)
	log.Printf("Finished with code %d\n", code)
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
	return filepath.Join(rootDir, script.Path, script.Version)
}
