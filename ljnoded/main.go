package main

import (
	"flag"
	consulkv "github.com/armon/consul-kv"
	"github.com/myfreeweb/gomaplog"
	"github.com/myfreeweb/lightjail/util"
	"path"
)

var options struct {
	consulAddr, ipIface string
	logJson             bool
}

var logger = gomaplog.StdoutLogger(gomaplog.DefaultTemplateFormatter)
var kvClient *consulkv.Client

type Process struct {
	S string
}

var processes map[string]*Process

func (process *Process) Start() {
	logger.Info("Starting process", gomaplog.Extras{"S": process.S})
}
func (process *Process) Stop() {
	logger.Info("Stopping process", gomaplog.Extras{"S": process.S})
}

func main() {
	util.MustHaveExecutables("mount", "umount", "devfs", "ljspawn", "route", "uname")
	parseOptions()
	defer logger.HandlePanic()
	logger.Host = "ljnoded@" + util.Hostname()
	kvConf := consulkv.DefaultConfig()
	kvConf.Address = options.consulAddr
	kvClient, _ = consulkv.NewClient(kvConf)
	logger.Info("Node started", gomaplog.Extras{"consul": options.consulAddr})
	processes = make(map[string]*Process)
	loop()
}

func parseOptions() {
	flag.StringVar(&options.consulAddr, "c", "127.0.0.1:8500", "host:port of the nearest Consul agent")
	flag.StringVar(&options.ipIface, "f", "default", "network interface for the jails")
	flag.BoolVar(&options.logJson, "j", false, "log in JSON (GELF 1.1) format")
	flag.Parse()
	if options.logJson {
		logger = gomaplog.StdoutLogger(gomaplog.DefaultJSONFormatter)
	}
	if options.ipIface == "default" {
		options.ipIface = util.DefaultIpIface()
	}
}

func loop() {
	kvNodePath := path.Join("lj/nodes", util.Hostname()+"/")
	meta, list, err := kvClient.List(kvNodePath)
	if err != nil {
		panic(err)
	}
	handleKV(list)
	for {
		meta, list, err = kvClient.WatchList(kvNodePath, meta.ModifyIndex)
		if err != nil {
			logger.Error("Error contacting Consul", gomaplog.Extras{"error": err})
			// sleep ?exponentially?
		}
		handleKV(list)
	}
}

func handleKV(pairs consulkv.KVPairs) {
	processesToKill := make(map[string]bool)
	for k, _ := range processes {
		processesToKill[k] = true
	}
	for _, v := range pairs {
		if _, ok := processes[v.Key]; ok {
			processesToKill[v.Key] = false
			newProc := Process{S: string(v.Value)}
			newProc.Start()
			processes[v.Key].Stop()
			processes[v.Key] = &newProc
		} else {
			newProc := Process{S: string(v.Value)}
			newProc.Start()
			processes[v.Key] = &newProc
		}
	}
	for k, _ := range processes {
		if v, _ := processesToKill[k]; v {
			processes[k].Stop()
			delete(processes, k)
		}
	}
}
