package main

import (
	"crypto/rand"
	"encoding/base32"
	"log"
	"os/exec"
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
		providesMatches := regexp.MustCompile(`^provides ([^\n# ]+).*`).FindStringSubmatch(line)
		if providesMatches != nil {
			script.Path = providesMatches[1]
		}
		versionMatches := regexp.MustCompile(`^version ([^\n# ]+).*`).FindStringSubmatch(line)
		if versionMatches != nil {
			script.Version = versionMatches[1]
		}
	}
	if script.WorldVersion == "" {
		script.WorldVersion = defaultWorldVersion()
		log.Printf("The Jailfile does not contain a 'world' directive, using default: %s", script.WorldVersion)
	}
	if script.Path == "" {
		log.Fatal("The Jailfile does not contain a 'provides' directive")
	}
	if script.Version == "" && overrideVersion == "" {
		script.Version = randomVersion()
		log.Printf("The Jailfile does not contain a 'version' directive, using random: %s", script.Version)
	}
	if overrideVersion != "" {
		script.Version = overrideVersion
	}
	return script
}

func defaultWorldVersion() string {
	ver, err := exec.Command("uname", "-r").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(ver))
}

func randomVersion() string {
	ver := make([]byte, 8)
	_, err := rand.Read(ver)
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(ver), "=")
}
