package main

import (
	"github.com/myfreeweb/lightjail/util"
	"log"
	"regexp"
	"strings"
)

func parseJailfile(src, overrideVersion string) *Script {
	fileParts := strings.SplitN(src, "\n---\n", 2)
	if len(fileParts) != 2 {
		log.Fatal("The Jailfile does not contain a '---' separator")
	}
	script := new(Script)
	script.Buildscript = fileParts[1]
	for _, line := range strings.Split(fileParts[0], "\n") {

		worldMatches := regexp.MustCompile(`^world ([^\n ]+).*`).FindStringSubmatch(line)
		if worldMatches != nil {
			script.WorldVersion = worldMatches[1]
		}

		nameMatches := regexp.MustCompile(`^name ([^\n# ]+).*`).FindStringSubmatch(line)
		if nameMatches != nil {
			script.Name = nameMatches[1]
		}

		fromMatches := regexp.MustCompile(`^from ([^\n# ]+) ([^\n# ]+).*`).FindStringSubmatch(line)
		if fromMatches != nil {
			script.From = &Overlay{Name: fromMatches[1], Version: fromMatches[2]}
		}

		versionMatches := regexp.MustCompile(`^version ([^\n# ]+).*`).FindStringSubmatch(line)
		if versionMatches != nil {
			script.Version = versionMatches[1]
		}

	}

	if script.WorldVersion == "" {
		script.WorldVersion = util.DefaultWorldVersion()
		log.Printf("The Jailfile does not contain a 'world' directive, using default: %s", script.WorldVersion)
	}

	if script.Version == "" && overrideVersion == "" {
		script.Version = util.RandomVersion()
		log.Printf("The Jailfile does not contain a 'version' directive, using random: %s", script.Version)
	}

	if overrideVersion != "" {
		script.Version = overrideVersion
	}

	return script
}
