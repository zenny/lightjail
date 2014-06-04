package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

        "github.com/myfreeweb/gomaplog"
        "github.com/myfreeweb/lightjail/util"
)

var logger = gomaplog.StdoutLogger(gomaplog.DefaultTemplateFormatter)

var options struct {
	overrideVersion, ipAddr, ipIface string
	memLimit                         int
	logJson                          bool
}

func main() {
	parseOptions()
	defer logger.HandlePanic()
	logger.Host = "ljbuild@" + util.Hostname()
	util.MustHaveExecutables("mount", "umount", "devfs", "ljspawn", "route", "uname", "cp")
	runJailfile(jailfilePath(flag.Args()))
}

func jailfilePath(args []string) string {
	jfPath := "Jailfile"
	if len(args) >= 1 {
		jfPath = args[0]
	}
	jfPath, err := filepath.Abs(jfPath)
	if err != nil {
		panic(err)
	}
	return jfPath
}

func parseOptions() {
	flag.StringVar(&options.overrideVersion, "v", "", "Override version")
	flag.StringVar(&options.ipAddr, "i", "192.168.1.240", "IPv4 address of the build jail")
	flag.StringVar(&options.ipIface, "f", "default", "network interface for the build jail")
	flag.IntVar(&options.memLimit, "m", 256, "virtual memory limit in megabytes")
	flag.BoolVar(&options.logJson, "j", false, "log in JSON (GELF 1.1) format")
	flag.Parse()
	if options.logJson {
		logger = gomaplog.StdoutLogger(gomaplog.DefaultJSONFormatter)
	}
	if options.ipIface == "default" {
		options.ipIface = util.DefaultIpIface()
	}
}

func runJailfile(path string) {
	tmpdir := util.MustTempDir("ljbuild-")
	jail := Buildjail{
		Script:     readAndParseJailfile(path, options.overrideVersion),
		MountPoint: tmpdir,
		Name:       filepath.Base(tmpdir),
		MemLimit:   options.memLimit,
		RootDir:    util.RootDir(),
	}
	jail.MustValidate()
	jail.MustMakeOverlayDir()
	defer syscall.Rmdir(jail.MountPoint)
	logger.Info("Build started", gomaplog.Extras{"name": jail.Script.Name, "version": jail.Script.Version})
	jail.CopyFiles()
	mounter := new(util.Mounter)
	defer mounter.Cleanup()
	jail.Mount(mounter)
	runner := new(util.Runner)
	startHandlingInterrupts(runner)
	exitCode := <-runner.Run(jail.Cmd(), "build")
	jail.Script.Overlay.Save(filepath.Join(jail.GetOverlayPath(), "overlay.json"))
	time.Sleep(300 * time.Millisecond) // Wait for jail removal, just in case
	logger.Info("Build finished", gomaplog.Extras{"status": exitCode})
}

func startHandlingInterrupts(runner *util.Runner) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		i := <-interrupt
		logger.Notice("Interrupted by a signal", gomaplog.Extras{"signal": i})
		runner.Cleanup()
		mounter.Cleanup()
	}()
}
