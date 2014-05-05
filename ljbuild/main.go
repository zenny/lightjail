package main

import (
	"flag"
	"github.com/myfreeweb/lightjail/util"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var overrideVersion, ipAddr, ipIface string

func main() {
	if _, err := exec.LookPath("ljspawn"); err != nil {
		log.Fatal(err)
	}
	if _, err := exec.LookPath("mount"); err != nil {
		log.Fatal(err)
	}
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
	script.RootDir = getRootDir()
	script.Validate()
	if err := os.MkdirAll(script.GetOverlayPath(), 0774); err != nil {
		log.Fatal(err)
	}
	mountPoint, err := ioutil.TempDir("", "ljbuild-")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Building %s version %s\n", script.Path, script.Version)
	mounter := new(util.Mounter)
	mounter.Mount("nullfs", "ro", script.GetWorldDir(), mountPoint)
	mounter.Mount("unionfs", "rw", script.GetOverlayPath(), mountPoint)
	ljspawnCmd := exec.Command("ljspawn", "-n", filepath.Base(mountPoint), "-i", ipAddr, "-f", ipIface, "-d", mountPoint, "-p", script.Buildscript)
	ljspawnCmd.Stdout = os.Stdout
	ljspawnCmd.Stderr = os.Stdout
	runner := new(util.Runner)
	handleInterrupts(runner)
	code := <-runner.Run(ljspawnCmd)
	time.Sleep(300 * time.Millisecond) // Wait for jail removal, just in case
	mounter.UnmountAll()
	syscall.Rmdir(mountPoint)
	log.Printf("Finished with code %d\n", code)
}

func handleInterrupts(runner *util.Runner) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-interrupt
		runner.Stop()
	}()
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
