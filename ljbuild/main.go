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
	overrideVersion, ipAddr, ipIface, ramLimitSoft, ramLimitHard string
}

func main() {
	util.CheckExecutables([]string{"mount", "umount", "devfs", "rctl", "ljspawn", "route", "uname", "cp"})
	parseOptions()
	args := flag.Args()
	var jfPath string
	if len(args) < 1 {
		jfPath = "Jailfile"
	} else {
		jfPath = args[0]
	}
	jfPath, err := filepath.Abs(jfPath)
	if err != nil {
		log.Fatal(err)
	}
	runJailfile(jfPath)
}

func parseOptions() {
	flag.StringVar(&options.overrideVersion, "v", "", "Override version")
	flag.StringVar(&options.ipAddr, "i", "", "IPv4 address of the build jail")
	flag.StringVar(&options.ipIface, "f", "", "network interface for the build jail")
	flag.StringVar(&options.ramLimitSoft, "r", "", "soft RAM limit (eg. 500m)")
	flag.StringVar(&options.ramLimitHard, "R", "", "hard RAM limit (eg. 512m)")
	flag.Parse()
	if options.ipAddr == "" {
		options.ipAddr = "192.168.1.240"
	}
	if options.ipIface == "" {
		options.ipIface = util.DefaultIpIface()
	}
	if options.ramLimitSoft == "" {
		options.ramLimitSoft = "500m"
	}
	if options.ramLimitHard == "" {
		options.ramLimitHard = "512m"
	}
}

func runJailfile(path string) {
	script := parseJailfile(readFile(path), options.overrideVersion)
	script.RootDir = util.RootDir()
	script.Validate()
	if err := os.MkdirAll(script.GetOverlayPath(), 0774); err != nil {
		log.Fatal(err)
	}
	mountPoint, err := ioutil.TempDir("", "ljbuild-")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Building %s version %s\n", script.Name, script.Version)
	if script.CopyDst != "" {
		if err := exec.Command("cp", "-R", filepath.Dir(path)+"/", filepath.Join(script.GetOverlayPath(), script.CopyDst)).Run(); err != nil {
			log.Fatal(err)
		}
	}
	mounter := new(util.Mounter)
	mounter.Mount("nullfs", "ro", script.GetWorldDir(), mountPoint)
	for _, path := range script.GetFromPaths() {
		mounter.Mount("unionfs", "ro", path, mountPoint)
	}
	mounter.Mount("unionfs", "rw", script.GetOverlayPath(), mountPoint)
	mounter.MountDev(filepath.Join(mountPoint, "dev"))
	jailName := filepath.Base(mountPoint)
	rctl := new(util.Rctl)
	rctl.LimitJailRam(jailName, options.ramLimitSoft, options.ramLimitHard)
	ljspawnCmd := exec.Command("ljspawn", "-n", jailName, "-i", options.ipAddr, "-f", options.ipIface, "-d", mountPoint, "-p", script.Buildscript)
	ljspawnCmd.Stdout = os.Stdout
	ljspawnCmd.Stderr = os.Stdout
	runner := new(util.Runner)
	handleInterrupts(runner)
	exitCode := <-runner.Run(ljspawnCmd)
	script.Overlay.Save(filepath.Join(script.GetOverlayPath(), "overlay.json"))
	time.Sleep(300 * time.Millisecond) // Wait for jail removal, just in case
	mounter.UnmountAll()
	rctl.RemoveAll()
	syscall.Rmdir(mountPoint)
	log.Printf("Finished with code %d\n", exitCode)
}

func handleInterrupts(runner *util.Runner) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-interrupt
		log.Print("Interrupted by a signal, stopping.\n")
		runner.Stop()
	}()
}

func readFile(path string) string {
	fileSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(fileSrc)
}
