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

var options struct {
	overrideVersion, ipAddr, ipIface string
}

func main() {
	util.CheckExecutables([]string{"mount", "umount", "ljspawn", "route", "uname"})
	parseOptions()
	args := flag.Args()
	var jfPath string
	if len(args) < 1 {
		jfPath = "Jailfile"
	} else {
		jfPath = args[0]
	}
	runJailfile(jfPath)
}

func parseOptions() {
	flag.StringVar(&options.overrideVersion, "v", "", "Override version")
	flag.StringVar(&options.ipAddr, "i", "", "IPv4 address of the build jail")
	flag.StringVar(&options.ipIface, "f", "", "network interface for the build jail")
	flag.Parse()
	if options.ipAddr == "" {
		options.ipAddr = "192.168.1.240"
	}
	if options.ipIface == "" {
		options.ipIface = util.DefaultIpIface()
	}
}

func runJailfile(path string) {
	script := parseJailfile(readJailfile(path), options.overrideVersion)
	script.RootDir = util.RootDir()
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
	ljspawnCmd := exec.Command("ljspawn", "-n", filepath.Base(mountPoint), "-i", options.ipAddr, "-f", options.ipIface, "-d", mountPoint, "-p", script.Buildscript)
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
