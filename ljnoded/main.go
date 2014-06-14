package main

import (
	"encoding/json"
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

func main() {
	util.MustHaveExecutables("mount", "umount", "devfs", "ljspawn", "route", "uname")
	parseOptions()
	// defer logger.HandlePanic()
	logger.Host = "ljnoded@" + util.Hostname()
	kvConf := consulkv.DefaultConfig()
	kvConf.Address = options.consulAddr
	kvClient, _ = consulkv.NewClient(kvConf)
	logger.Info("Node started", gomaplog.Extras{"consul": options.consulAddr})
	procNode := NewProcNode()
	procNode.StartHandlingInterrupts()
	loop(procNode)
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

func loop(node *ProcNode) {
	kvNodePath := path.Join("lj/nodes", util.Hostname(), "run")
	meta, list, err := kvClient.List(kvNodePath)
	if err != nil {
		logger.Error("Error contacting Consul", gomaplog.Extras{"error": err})
		panic(err)
	}
	handleKV(list, node)
	for {
		meta, list, err = kvClient.WatchList(kvNodePath, meta.ModifyIndex)
		if err != nil {
			logger.Error("Error contacting Consul", gomaplog.Extras{"error": err})
			// sleep ?exponentially?
		}
		handleKV(list, node)
	}
}

func handleKV(pairs consulkv.KVPairs, node *ProcNode) {
	newProcesses := make(map[string]*Process)
	for _, v := range pairs {
		var proc Process
		err := json.Unmarshal(v.Value, &proc)
		if err != nil {
			logger.Warning("Invalid JSON found, ignoring", gomaplog.Extras{"key": v.Key, "consul": options.consulAddr})
		} else {
			newProcesses[v.Key] = &proc
		}
	}
	node.HandleProcesses(newProcesses)
}

func writeProcess(key string, proc *Process) error {
	kvPath := path.Join("lj/nodes", util.Hostname(), "run", key)
	bytes, err := json.Marshal(proc)
	if err != nil {
		logger.Error("Error encoding process as JSON", gomaplog.Extras{"error": err, "key": kvPath})
		return err
	}
	err = kvClient.Put(key, bytes, 0)
	if err != nil {
		logger.Error("Error writing process info to Consul", gomaplog.Extras{"error": err, "key": kvPath, "consul": options.consulAddr})
		return err
	}
	return nil
}
