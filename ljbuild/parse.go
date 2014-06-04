package main

import (
	"regexp"
	"strings"

        "github.com/myfreeweb/gomaplog"
        "github.com/myfreeweb/lightjail/util"
)

func parseJailfile(src, overrideVersion string) *Script {
	fileParts := strings.SplitN(src, "\n#!", 2)
	if len(fileParts) != 2 {
		panic("The Jailfile does not contain '#!'")
	}
	script := new(Script)
	script.Buildscript = "#!" + fileParts[1]
	for _, line := range strings.Split(fileParts[0], "\n") {

		worldMatches := regexp.MustCompile(`^world ([^\n ]+).*`).FindStringSubmatch(line)
		if worldMatches != nil {
			script.WorldVersion = worldMatches[1]
		}

		nameMatches := regexp.MustCompile(`^name ([^\n# ]+).*`).FindStringSubmatch(line)
		if nameMatches != nil {
			script.Name = nameMatches[1]
		}

		fromMatches := regexp.MustCompile(`^from ([^\n# ]+) ([^\n#]+).*`).FindStringSubmatch(line)
		if fromMatches != nil {
			script.From = &util.Overlay{Name: fromMatches[1], Version: fromMatches[2]}
		}

		versionMatches := regexp.MustCompile(`^version ([^\n# ]+).*`).FindStringSubmatch(line)
		if versionMatches != nil {
			script.Version = versionMatches[1]
		}

		copyDstMatches := regexp.MustCompile(`^copy ([^\n# ]+).*`).FindStringSubmatch(line)
		if copyDstMatches != nil {
			script.CopyDst = copyDstMatches[1]
		}

	}

	if script.WorldVersion == "" {
		script.WorldVersion = util.DefaultWorldVersion()
		logger.Info("The Jailfile does not contain a 'world' directive, using default", gomaplog.Extras{"world": script.WorldVersion, "name": script.Name})
	}

	if script.Version == "" && overrideVersion == "" {
		script.Version = util.RandomVersion()
		logger.Notice("The Jailfile does not contain a 'version' directive, using random", gomaplog.Extras{"version": script.Version, "name": script.Name})
	}

	if overrideVersion != "" {
		script.Version = overrideVersion
	}

	return script
}

func readAndParseJailfile(path, overrideVersion string) *Script {
	script := parseJailfile(util.MustReadFileStr(path), overrideVersion)
	script.ScriptPath = path
	return script
}
